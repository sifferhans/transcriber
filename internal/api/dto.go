package api

import (
	"time"

	"transcriber/internal/jobs"
)

type TranscribeInput struct {
	Path       string `json:"path"`
	Language   string `json:"language"`
	Format     string `json:"format"`
	Callback   string `json:"callback,omitempty"`
	OutputPath string `json:"output_path"`
	Priority   int    `json:"priority,omitempty"`
	Model      string `json:"model,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
}

type TranscribeJob struct {
	ID           string   `json:"id"`
	Path         string   `json:"path"`
	Language     string   `json:"language"`
	OutputFormat string   `json:"format"`
	OutputPath   string   `json:"output_path"`
	Progress     int      `json:"progress"`
	Status       string   `json:"status"`
	Result       string   `json:"result"`
	Results      []string `json:"results,omitempty"`
	Callback     string   `json:"callback"`
	Model        string   `json:"model"`
	Prompt       string   `json:"prompt,omitempty"`
	Duration     string   `json:"duration"`
	Priority     int      `json:"priority"`
	Error        string   `json:"error,omitempty"`
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
		Results:      j.Results,
		Callback:     j.Callback,
		Model:        j.Model,
		Prompt:       j.Prompt,
		Duration:     formatDuration(j.Duration),
		Priority:     j.Priority,
		Error:        j.Error,
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	return "duration: " + d.String()
}
