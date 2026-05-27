package callback

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Notifier delivers webhook callbacks on a worker pool with exponential backoff.
type Notifier struct {
	queue   chan task
	client  *http.Client
	workers int
	wg      sync.WaitGroup
}

type task struct {
	url  string
	body []byte
}

func NewNotifier(workers, queueSize int) *Notifier {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 64
	}
	return &Notifier{
		queue:   make(chan task, queueSize),
		workers: workers,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (n *Notifier) Start(ctx context.Context) {
	for i := 0; i < n.workers; i++ {
		n.wg.Add(1)
		go n.run(ctx)
	}
}

// Shutdown closes the queue and waits for in-flight callbacks to drain.
func (n *Notifier) Shutdown() {
	close(n.queue)
	n.wg.Wait()
}

// Enqueue queues a JSON body to be POSTed to url; returns an error if the queue is full.
func (n *Notifier) Enqueue(url string, body []byte) error {
	if url == "" {
		return nil
	}
	select {
	case n.queue <- task{url: url, body: body}:
		return nil
	default:
		return errors.New("callback queue full")
	}
}

func (n *Notifier) run(ctx context.Context) {
	defer n.wg.Done()
	for {
		select {
		case t, ok := <-n.queue:
			if !ok {
				return
			}
			n.deliver(ctx, t)
		case <-ctx.Done():
			return
		}
	}
}

func (n *Notifier) deliver(ctx context.Context, t task) {
	backoff := time.Second
	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(t.body))
		if err != nil {
			slog.Warn("callback request build failed", "url", t.url, "err", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := n.client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return
			}
			err = errors.New(resp.Status)
		}
		slog.Warn("callback delivery failed", "url", t.url, "attempt", attempt, "err", err)
		if attempt == maxAttempts {
			return
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		backoff *= 2
	}
}
