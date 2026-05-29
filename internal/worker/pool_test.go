package worker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"transcriber/internal/jobs"
	"transcriber/internal/transcriber"
	"transcriber/internal/transcriber/stub"
)

// stub.Adapter sleeps ~2 s before returning, so any timeout under that should
// reliably mark the job FAILED with error "timeout".
func TestPool_Timeouts(t *testing.T) {
	cases := []struct {
		name           string
		defaultTimeout time.Duration
		jobTimeout     time.Duration
	}{
		{"per-job override applies", 0, 300 * time.Millisecond},
		{"pool default applies when job has none", 200 * time.Millisecond, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runJobAndWait(t, tc.defaultTimeout, tc.jobTimeout)
			if got.Status != jobs.StatusFailed {
				t.Fatalf("status: got %q, want FAILED", got.Status)
			}
			if got.Error != "timeout" {
				t.Fatalf("error: got %q, want %q", got.Error, "timeout")
			}
		})
	}
}

func runJobAndWait(t *testing.T, defaultTimeout, jobTimeout time.Duration) jobs.Job {
	t.Helper()
	tmp := t.TempDir()

	store := jobs.NewStore(0)
	queue := jobs.NewQueue()
	t.Cleanup(queue.Close)

	registry := transcriber.NewRegistry("stub")
	registry.Register(stub.New("stub", "Stub"))

	pool := New(1, store, queue, registry, nil, func(j jobs.Job) any { return j }, defaultTimeout)
	pool.Start(t.Context())

	now := time.Now()
	job := jobs.Job{
		ID:         t.Name(),
		Path:       "/dev/null",
		OutputPath: filepath.Join(tmp, "out"),
		Model:      "stub",
		Timeout:    jobTimeout,
		Status:     jobs.StatusPending,
		CreatedAt:  now,
	}
	if err := os.MkdirAll(job.OutputPath, 0o755); err != nil {
		t.Fatal(err)
	}
	store.Create(job)
	queue.Push(job.ID, 1, now)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if j, ok := store.Get(job.ID); ok && isTerminal(j.Status) {
			return j
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("job did not reach a terminal state")
	return jobs.Job{}
}

func isTerminal(s string) bool {
	return s == jobs.StatusCompleted || s == jobs.StatusFailed || s == jobs.StatusCanceled
}
