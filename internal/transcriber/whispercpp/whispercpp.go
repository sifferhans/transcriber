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
	ID        string
	Binary    string
	ModelFile string
	Threads   int
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
	if a.cfg.ModelFile == "" {
		return nil, fmt.Errorf("whispercpp: model_file not configured")
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}
	outPrefix := filepath.Join(req.OutputDir, "whispercpp_out")

	args := []string{
		"-m", a.cfg.ModelFile,
		"-f", req.InputPath,
		"-of", outPrefix,
		"-oj",
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

// whisper.cpp `--output-json` shape (timestamps in ms via offsets.from/to).
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
	Text string `json:"text"`
}

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
		})
		sb.WriteString(text)
		sb.WriteByte(' ')
	}
	t.Text = strings.TrimSpace(sb.String())
	return t, nil
}
