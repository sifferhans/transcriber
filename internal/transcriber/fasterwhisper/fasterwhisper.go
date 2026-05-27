// Package fasterwhisper adapts the `whisper-ctranslate2` CLI (the
// faster-whisper distribution) to the Transcriber interface. The CLI writes a
// JSON file named after the input audio file; we parse it into the unified
// Transcription shape.
package fasterwhisper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"transcriber/internal/transcriber"
)

type Config struct {
	ID          string
	Binary      string
	Model       string // model name or path, e.g. "large-v3"
	ComputeType string // float16, int8_float16, int8, ...
	Device      string // cuda | cpu | auto
}

type Adapter struct {
	cfg Config
}

func New(cfg Config) *Adapter {
	if cfg.ID == "" {
		cfg.ID = "faster-whisper"
	}
	return &Adapter{cfg: cfg}
}

func (a *Adapter) ID() string { return a.cfg.ID }

func (a *Adapter) Name() string {
	if a.cfg.Model == "" {
		return "faster-whisper"
	}
	return "faster-whisper (" + a.cfg.Model + ")"
}

func (a *Adapter) Transcribe(ctx context.Context, req transcriber.Request, onProgress transcriber.ProgressFunc) (*transcriber.Result, error) {
	if a.cfg.Binary == "" {
		return nil, fmt.Errorf("fasterwhisper: binary not configured")
	}
	if a.cfg.Model == "" {
		return nil, fmt.Errorf("fasterwhisper: model not configured")
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}

	args := []string{
		req.InputPath,
		"--model", a.cfg.Model,
		"--output_dir", req.OutputDir,
		"--output_format", "json",
		"--word_timestamps", "true",
	}
	if a.cfg.ComputeType != "" {
		args = append(args, "--compute_type", a.cfg.ComputeType)
	}
	if a.cfg.Device != "" {
		args = append(args, "--device", a.cfg.Device)
	}
	if req.Language != "" && req.Language != "auto" {
		args = append(args, "--language", req.Language)
	}

	// whisper-ctranslate2 does not emit a numeric progress signal we can rely
	// on, so we just drain its output. Progress stays at 0 until completion.
	// onProgress is intentionally unused.
	_ = onProgress

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
		return nil, fmt.Errorf("fasterwhisper start: %w", err)
	}
	go io.Copy(io.Discard, stdout)
	go io.Copy(io.Discard, stderr)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("fasterwhisper exit: %w", err)
	}

	base := strings.TrimSuffix(filepath.Base(req.InputPath), filepath.Ext(req.InputPath))
	outPath := filepath.Join(req.OutputDir, base+".json")
	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("fasterwhisper output: %w", err)
	}
	tr, err := parseJSON(data)
	if err != nil {
		return nil, err
	}
	return &transcriber.Result{
		Transcription: tr,
		ModelUsed:     a.cfg.ID,
		Duration:      time.Since(start),
	}, nil
}

// faster-whisper / whisper-ctranslate2 JSON shape (closely matches OpenAI's,
// but uses "word" instead of "text" in word entries and "probability" for
// confidence).
type rawOutput struct {
	Text     string       `json:"text"`
	Language string       `json:"language"`
	Segments []rawSegment `json:"segments"`
}

type rawSegment struct {
	ID               int       `json:"id"`
	Start            float64   `json:"start"`
	End              float64   `json:"end"`
	Text             string    `json:"text"`
	Tokens           []int     `json:"tokens"`
	Temperature      float64   `json:"temperature"`
	AvgLogprob       float64   `json:"avg_logprob"`
	CompressionRatio float64   `json:"compression_ratio"`
	NoSpeechProb     float64   `json:"no_speech_prob"`
	Words            []rawWord `json:"words"`
}

type rawWord struct {
	Word        string  `json:"word"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	Probability float64 `json:"probability"`
}

func parseJSON(data []byte) (*transcriber.Transcription, error) {
	var raw rawOutput
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("fasterwhisper parse: %w", err)
	}
	t := &transcriber.Transcription{
		Text:     strings.TrimSpace(raw.Text),
		Language: raw.Language,
		Segments: make([]transcriber.Segment, 0, len(raw.Segments)),
	}
	for _, s := range raw.Segments {
		seg := transcriber.Segment{
			ID:               s.ID,
			Start:            s.Start,
			End:              s.End,
			Text:             s.Text,
			Tokens:           s.Tokens,
			Temperature:      s.Temperature,
			AvgLogprob:       s.AvgLogprob,
			CompressionRatio: s.CompressionRatio,
			NoSpeechProb:     s.NoSpeechProb,
		}
		for _, w := range s.Words {
			seg.Words = append(seg.Words, transcriber.Word{
				Text:  w.Word,
				Start: w.Start,
				End:   w.End,
			})
		}
		t.Segments = append(t.Segments, seg)
	}
	return t, nil
}
