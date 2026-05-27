// Package whispercpp adapts the whisper.cpp CLI (the `main` / `whisper-cli`
// binary) to the Transcriber interface. The binary writes a JSON file we
// then parse into the unified Transcription shape.
package whispercpp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"transcriber/internal/transcriber"
)

type Config struct {
	ID     string
	Binary string
	// ModelFile is a local path to the ggml model file. If set, it wins over
	// ResolveModel — useful for operator-pinned deployments and tests.
	ModelFile string
	// ResolveModel is called on first Transcribe when ModelFile is empty.
	// Typically wired to hfcache.Cache.Get so the model is downloaded from
	// Hugging Face on demand and cached on disk. Keeping this as a hook
	// (rather than a direct hfcache dep) decouples the adapter from any
	// specific resolver and keeps tests cheap.
	ResolveModel func(ctx context.Context) (string, error)
	Threads      int
}

type Adapter struct {
	cfg Config
}

func New(cfg Config) *Adapter {
	if cfg.ID == "" {
		cfg.ID = "whisper-cpp"
	}
	if cfg.Threads <= 0 {
		cfg.Threads = 4
	}
	return &Adapter{cfg: cfg}
}

func (a *Adapter) ID() string { return a.cfg.ID }

func (a *Adapter) Name() string {
	if a.cfg.ModelFile == "" {
		return "whisper.cpp"
	}
	return "whisper.cpp (" + filepath.Base(a.cfg.ModelFile) + ")"
}

var progressRe = regexp.MustCompile(`progress\s*=\s*(\d+)`)

func (a *Adapter) Transcribe(ctx context.Context, req transcriber.Request, onProgress transcriber.ProgressFunc) (*transcriber.Result, error) {
	if a.cfg.Binary == "" {
		return nil, fmt.Errorf("whispercpp: binary not configured")
	}
	modelPath := a.cfg.ModelFile
	if modelPath == "" {
		if a.cfg.ResolveModel == nil {
			return nil, fmt.Errorf("whispercpp: no model_file and no resolver configured")
		}
		p, err := a.cfg.ResolveModel(ctx)
		if err != nil {
			return nil, fmt.Errorf("whispercpp resolve model: %w", err)
		}
		modelPath = p
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}
	outPrefix := filepath.Join(req.OutputDir, "whispercpp_out")

	args := []string{
		"-m", modelPath,
		"-f", req.InputPath,
		"-of", outPrefix,
		"-ojf",
		"-t", strconv.Itoa(a.cfg.Threads),
		"-pp",
	}
	if req.Language != "" && req.Language != "auto" {
		args = append(args, "--language", req.Language)
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, a.cfg.Binary, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("whispercpp start: %w", err)
	}

	capture := captureStderr(stderr, onProgress)
	go io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		tail := capture.wait()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if tail != "" {
			return nil, fmt.Errorf("whispercpp exit: %w: %s", err, tail)
		}
		return nil, fmt.Errorf("whispercpp exit: %w", err)
	}

	data, err := os.ReadFile(outPrefix + ".json")
	if err != nil {
		return nil, fmt.Errorf("whispercpp output: %w", err)
	}
	tr, err := parseJSON(data, req.Language)
	if err != nil {
		return nil, err
	}
	return &transcriber.Result{
		Transcription: tr,
		ModelUsed:     a.cfg.ID,
		Duration:      time.Since(start),
	}, nil
}

// stderrCapture reads whisper-cli's stderr, dispatches progress lines to
// onProgress, and buffers the tail of non-progress lines so they can be
// included in the exit error when the process fails.
type stderrCapture struct {
	done chan struct{}
	tail []string
}

func captureStderr(r io.Reader, onProgress transcriber.ProgressFunc) *stderrCapture {
	c := &stderrCapture{done: make(chan struct{})}
	go func() {
		defer close(c.done)
		const maxTail = 20
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			if m := progressRe.FindStringSubmatch(line); m != nil {
				if onProgress != nil {
					if pct, err := strconv.Atoi(m[1]); err == nil {
						onProgress(float64(pct) / 100.0)
					}
				}
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			c.tail = append(c.tail, line)
			if len(c.tail) > maxTail {
				c.tail = c.tail[len(c.tail)-maxTail:]
			}
		}
	}()
	return c
}

func (c *stderrCapture) wait() string {
	<-c.done
	return strings.Join(c.tail, "\n")
}

// whisper.cpp `--output-json-full` shape: each transcription entry contains
// the text plus per-token timestamps (ms via offsets.from/to). Tokens are
// BPE pieces — words start at a token whose text begins with a space.
type rawOutput struct {
	Result struct {
		Language string `json:"language"`
	} `json:"result"`
	Transcription []rawSegment `json:"transcription"`
}

type rawSegment struct {
	Offsets struct {
		From int `json:"from"` // milliseconds
		To   int `json:"to"`
	} `json:"offsets"`
	Text   string     `json:"text"`
	Tokens []rawToken `json:"tokens"`
}

type rawToken struct {
	Text    string `json:"text"`
	Offsets struct {
		From int `json:"from"`
		To   int `json:"to"`
	} `json:"offsets"`
	ID int `json:"id"`
}

// Whisper's vocab puts the first special token at id 50256 ([_BEG_]/<|endoftext|>);
// anything at or above this is a control token that shouldn't appear in words.
const firstSpecialTokenID = 50256

func parseJSON(data []byte, fallbackLang string) (*transcriber.Transcription, error) {
	var raw rawOutput
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("whispercpp parse: %w", err)
	}
	t := &transcriber.Transcription{
		Language: raw.Result.Language,
		Segments: make([]transcriber.Segment, 0, len(raw.Transcription)),
	}
	if t.Language == "" {
		t.Language = fallbackLang
	}
	var sb strings.Builder
	for i, s := range raw.Transcription {
		text := strings.TrimSpace(s.Text)
		t.Segments = append(t.Segments, transcriber.Segment{
			ID:    i,
			Start: float64(s.Offsets.From) / 1000.0,
			End:   float64(s.Offsets.To) / 1000.0,
			Text:  text,
			Words: tokensToWords(s.Tokens),
		})
		sb.WriteString(text)
		sb.WriteByte(' ')
	}
	t.Text = strings.TrimSpace(sb.String())
	return t, nil
}

// tokensToWords groups BPE tokens into words. A new word begins when a
// token's text starts with a space; subsequent tokens (sub-word pieces or
// trailing punctuation) attach to the current word.
func tokensToWords(tokens []rawToken) []transcriber.Word {
	if len(tokens) == 0 {
		return nil
	}
	var words []transcriber.Word
	var cur *transcriber.Word
	flush := func() {
		if cur == nil {
			return
		}
		cur.Text = strings.TrimSpace(cur.Text)
		if cur.Text != "" {
			words = append(words, *cur)
		}
		cur = nil
	}
	for _, tok := range tokens {
		if tok.ID >= firstSpecialTokenID {
			continue
		}
		startsWord := cur == nil || strings.HasPrefix(tok.Text, " ")
		if startsWord {
			flush()
			start := float64(tok.Offsets.From) / 1000.0
			end := float64(tok.Offsets.To) / 1000.0
			cur = &transcriber.Word{Text: tok.Text, Start: start, End: end}
		} else {
			cur.Text += tok.Text
			if end := float64(tok.Offsets.To) / 1000.0; end > cur.End {
				cur.End = end
			}
		}
	}
	flush()
	return words
}
