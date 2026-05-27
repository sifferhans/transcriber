package jobs

import "time"

const (
	StatusPending   = "QUEUED"
	StatusRunning   = "RUNNING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
	StatusCanceled  = "CANCELED"
)

type Job struct {
	ID         string
	Path       string
	Language   string
	Format     string
	OutputPath string
	Priority   int
	Callback   string
	Model      string
	Prompt     string

	Status   string
	Progress int
	Result   string
	Error    string

	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
}
