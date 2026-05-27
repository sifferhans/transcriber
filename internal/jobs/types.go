package jobs

import "time"

// Status constants kept uppercase to match the existing Python API contract.
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

	Status   string
	Progress int    // 0-100
	Result   string // path to primary output file once COMPLETED
	Error    string

	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
}
