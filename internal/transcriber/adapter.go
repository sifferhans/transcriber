package transcriber

import (
	"context"
	"time"
)

// Transcription is the unified transcript shape returned by every adapter.
// It mirrors the OpenAI Whisper JSON layout used in the existing Python API.
type Transcription struct {
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Language string    `json:"language"`
}

type Segment struct {
	Text             string  `json:"text"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Words            []Word  `json:"words,omitempty"`
	ID               int     `json:"id"`
	Seek             int     `json:"seek,omitempty"`
	Tokens           []int   `json:"tokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	AvgLogprob       float64 `json:"avg_logprob,omitempty"`
	CompressionRatio float64 `json:"compression_ratio,omitempty"`
	NoSpeechProb     float64 `json:"no_speech_prob,omitempty"`
}

type Word struct {
	Text       string  `json:"text"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Request struct {
	InputPath string
	Language  string
	OutputDir string
	Options   map[string]any
}

type Result struct {
	Transcription *Transcription
	ModelUsed     string
	Duration      time.Duration
}

type ProgressFunc func(progress float64)

type Transcriber interface {
	ID() string
	Name() string
	Transcribe(ctx context.Context, req Request, onProgress ProgressFunc) (*Result, error)
}

type Registry struct {
	adapters map[string]Transcriber
	def      string
}

func NewRegistry(defaultID string) *Registry {
	return &Registry{adapters: map[string]Transcriber{}, def: defaultID}
}

func (r *Registry) Register(t Transcriber) {
	r.adapters[t.ID()] = t
}

func (r *Registry) Get(id string) (Transcriber, bool) {
	t, ok := r.adapters[id]
	return t, ok
}

func (r *Registry) Default() (Transcriber, bool) {
	return r.Get(r.def)
}

func (r *Registry) DefaultID() string {
	return r.def
}

func (r *Registry) List() []Transcriber {
	out := make([]Transcriber, 0, len(r.adapters))
	for _, t := range r.adapters {
		out = append(out, t)
	}
	return out
}
