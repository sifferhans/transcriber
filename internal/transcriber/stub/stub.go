// Package stub provides a fake Transcriber used for local development and
// integration tests when no real ASR backend is installed. It emits a fixed
// transcript after simulating ~2s of work with progress updates.
package stub

import (
	"context"
	"time"

	"github.com/bcc-code/transcriber/internal/transcriber"
)

type Adapter struct {
	id   string
	name string
}

func New(id, name string) *Adapter {
	if id == "" {
		id = "stub"
	}
	if name == "" {
		name = "Stub: " + id
	}
	return &Adapter{id: id, name: name}
}

func (a *Adapter) ID() string   { return a.id }
func (a *Adapter) Name() string { return a.name }

func (a *Adapter) Transcribe(ctx context.Context, req transcriber.Request, onProgress transcriber.ProgressFunc) (*transcriber.Result, error) {
	start := time.Now()
	const steps = 10
	for i := 1; i <= steps; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
		if onProgress != nil {
			onProgress(float64(i) / float64(steps))
		}
	}

	lang := req.Language
	if lang == "" || lang == "auto" {
		lang = "en"
	}

	t := &transcriber.Transcription{
		Language: lang,
		Text:     "This is a stub transcription.",
		Segments: []transcriber.Segment{
			{
				ID:    0,
				Start: 0.0,
				End:   2.0,
				Text:  "This is a stub transcription.",
				Words: []transcriber.Word{
					{Text: "This", Start: 0.0, End: 0.5, Confidence: 0.99},
					{Text: "is", Start: 0.5, End: 0.7, Confidence: 0.99},
					{Text: "a", Start: 0.7, End: 0.9, Confidence: 0.99},
					{Text: "stub", Start: 0.9, End: 1.3, Confidence: 0.95},
					{Text: "transcription.", Start: 1.3, End: 2.0, Confidence: 0.92},
				},
			},
		},
	}
	return &transcriber.Result{
		Transcription: t,
		ModelUsed:     a.id,
		Duration:      time.Since(start),
	}, nil
}
