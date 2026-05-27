package jobs

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// Queue is a goroutine-safe priority queue: higher Priority first, FIFO within ties.
type Queue struct {
	mu     sync.Mutex
	heap   itemHeap
	signal chan struct{} // buffered size 1, coalesces "something may be in the queue"
	done   chan struct{}
}

func NewQueue() *Queue {
	return &Queue{
		signal: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}
}

func (q *Queue) Push(id string, priority int, createdAt time.Time) {
	q.mu.Lock()
	heap.Push(&q.heap, &queueItem{id: id, priority: priority, createdAt: createdAt})
	q.mu.Unlock()
	q.notify()
}

func (q *Queue) Pop(ctx context.Context) (string, bool) {
	for {
		q.mu.Lock()
		if q.heap.Len() > 0 {
			item := heap.Pop(&q.heap).(*queueItem)
			more := q.heap.Len() > 0
			q.mu.Unlock()
			// Re-signal so other waiters wake up: the coalescing signal would otherwise wake only one.
			if more {
				q.notify()
			}
			return item.id, true
		}
		q.mu.Unlock()

		select {
		case <-q.signal:
		case <-q.done:
			return "", false
		case <-ctx.Done():
			return "", false
		}
	}
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.heap.Len()
}

func (q *Queue) Close() {
	select {
	case <-q.done:
	default:
		close(q.done)
	}
}

func (q *Queue) notify() {
	select {
	case q.signal <- struct{}{}:
	default:
	}
}

type queueItem struct {
	id        string
	priority  int
	createdAt time.Time
	index     int
}

type itemHeap []*queueItem

func (h itemHeap) Len() int { return len(h) }
func (h itemHeap) Less(i, j int) bool {
	if h[i].priority != h[j].priority {
		return h[i].priority > h[j].priority
	}
	return h[i].createdAt.Before(h[j].createdAt)
}
func (h itemHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *itemHeap) Push(x any) {
	item := x.(*queueItem)
	item.index = len(*h)
	*h = append(*h, item)
}
func (h *itemHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}
