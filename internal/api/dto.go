package api

import (
	"time"

	"transcriber/internal/jobs"
)

// TranscribeInput matches the request body shape of the existing Python API.
// `Model` is an additive optional field — existing clients that omit it get
// the server's default.
type TranscribeInput struct {
	Path       string `json:"path"`
	Language   string `json:"language"`
	Format     string `json:"format"`
	Callback   string `json:"callback,omitempty"`
	OutputPath string `json:"output_path"`
	Priority   int    `json:"priority,omitempty"`
	Model      string `json:"model,omitempty"`
}

// TranscribeJob mirrors the response shape of the existing Python API so
// existing callers (e.g. the Temporal activity that drives transcription)
// can talk to this server unchanged.
type TranscribeJob struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	Language     string `json:"language"`
	OutputFormat string `json:"format"`
	OutputPath   string `json:"output_path"`
	Progress     int    `json:"progress"`
	Status       string `json:"status"`
	Result       string `json:"result"`
	Callback     string `json:"callback"`
	Model        string `json:"model"`
	Duration     string `json:"duration"`
	Priority     int    `json:"priority"`
	Error        string `json:"error,omitempty"`
}

func ToDTO(j jobs.Job) TranscribeJob {
	return TranscribeJob{
		ID:           j.ID,
		Path:         j.Path,
		Language:     j.Language,
		OutputFormat: j.Format,
		OutputPath:   j.OutputPath,
		Progress:     j.Progress,
		Status:       j.Status,
		Result:       j.Result,
		Callback:     j.Callback,
		Model:        j.Model,
		Duration:     formatDuration(j.Duration),
		Priority:     j.Priority,
		Error:        j.Error,
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	return d.String()
}
