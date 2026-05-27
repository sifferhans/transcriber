package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"transcriber/internal/callback"
	"transcriber/internal/formats"
	"transcriber/internal/jobs"
	"transcriber/internal/transcriber"
)

// Pool is a fixed-size worker pool that runs jobs through the configured adapter.
type Pool struct {
	workers  int
	store    *jobs.Store
	queue    *jobs.Queue
	registry *transcriber.Registry
	notifier *callback.Notifier
	dtoFn    func(jobs.Job) any
	wg       sync.WaitGroup
}

func New(
	workers int,
	store *jobs.Store,
	queue *jobs.Queue,
	registry *transcriber.Registry,
	notifier *callback.Notifier,
	dtoFn func(jobs.Job) any,
) *Pool {
	if workers < 1 {
		workers = 1
	}
	return &Pool{
		workers:  workers,
		store:    store,
		queue:    queue,
		registry: registry,
		notifier: notifier,
		dtoFn:    dtoFn,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *Pool) Wait() { p.wg.Wait() }

func (p *Pool) worker(ctx context.Context, n int) {
	defer p.wg.Done()
	log := slog.With("worker", n)
	for {
		id, ok := p.queue.Pop(ctx)
		if !ok {
			log.Debug("worker exiting")
			return
		}
		p.runJob(ctx, id, log)
	}
}

func (p *Pool) runJob(parent context.Context, id string, log *slog.Logger) {
	job, ok := p.store.Get(id)
	if !ok {
		log.Warn("job vanished from store", "id", id)
		return
	}
	if job.Status == jobs.StatusCanceled {
		return
	}

	adapter, err := p.pickAdapter(job)
	if err != nil {
		p.markFailed(id, err)
		p.fireCallback(id)
		return
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	p.store.SetCancel(id, cancel)
	defer p.store.ClearCancel(id)

	startedAt := time.Now()
	p.store.Update(id, func(j *jobs.Job) {
		j.Status = jobs.StatusRunning
		j.StartedAt = startedAt
	})
	log.Info("job started", "id", id, "model", adapter.ID(), "path", job.Path)

	onProgress := func(progress float64) {
		pct := max(0, min(100, int(progress*100)))
		p.store.Update(id, func(j *jobs.Job) {
			j.Progress = max(j.Progress, pct)
		})
	}

	req := transcriber.Request{
		InputPath: job.Path,
		Language:  job.Language,
		OutputDir: job.OutputPath,
	}

	res, err := adapter.Transcribe(ctx, req, onProgress)
	if err != nil {
		if errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled {
			p.markCanceled(id)
			log.Info("job canceled", "id", id)
		} else {
			p.markFailed(id, err)
			log.Error("job failed", "id", id, "err", err)
		}
		p.fireCallback(id)
		return
	}

	wantFormats := formats.Parse(job.Format)
	basename := filepath.Base(job.Path)
	primary := ""
	for _, f := range wantFormats {
		path, err := formats.Write(f, res.Transcription, job.OutputPath, basename)
		if err != nil {
			p.markFailed(id, fmt.Errorf("write %s: %w", f, err))
			p.fireCallback(id)
			return
		}
		if f == formats.JSON || primary == "" {
			primary = path
		}
	}

	endedAt := time.Now()
	p.store.Update(id, func(j *jobs.Job) {
		j.Status = jobs.StatusCompleted
		j.Progress = 100
		j.EndedAt = endedAt
		j.Duration = endedAt.Sub(startedAt)
		j.Result = primary
	})
	log.Info("job completed", "id", id, "duration", endedAt.Sub(startedAt))
	p.fireCallback(id)
}

func (p *Pool) pickAdapter(j jobs.Job) (transcriber.Transcriber, error) {
	if j.Model != "" {
		a, ok := p.registry.Get(j.Model)
		if !ok {
			return nil, fmt.Errorf("unknown model: %s", j.Model)
		}
		return a, nil
	}
	a, ok := p.registry.Default()
	if !ok {
		return nil, errors.New("no default model configured")
	}
	return a, nil
}

func (p *Pool) markFailed(id string, err error) {
	p.store.Update(id, func(j *jobs.Job) {
		j.Status = jobs.StatusFailed
		j.Error = err.Error()
		j.EndedAt = time.Now()
		if !j.StartedAt.IsZero() {
			j.Duration = j.EndedAt.Sub(j.StartedAt)
		}
	})
}

func (p *Pool) markCanceled(id string) {
	p.store.Update(id, func(j *jobs.Job) {
		j.Status = jobs.StatusCanceled
		j.EndedAt = time.Now()
		if !j.StartedAt.IsZero() {
			j.Duration = j.EndedAt.Sub(j.StartedAt)
		}
	})
}

func (p *Pool) fireCallback(id string) {
	if p.notifier == nil {
		return
	}
	job, ok := p.store.Get(id)
	if !ok || job.Callback == "" {
		return
	}
	body, err := json.Marshal(p.dtoFn(job))
	if err != nil {
		slog.Warn("callback marshal failed", "id", id, "err", err)
		return
	}
	if err := p.notifier.Enqueue(job.Callback, body); err != nil {
		slog.Warn("callback enqueue failed", "id", id, "err", err)
	}
}
