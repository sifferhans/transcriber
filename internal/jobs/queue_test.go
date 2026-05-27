package jobs

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func popOrFail(t *testing.T, q *Queue) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	id, ok := q.Pop(ctx)
	if !ok {
		t.Fatal("expected Pop to return an item")
	}
	return id
}

func TestQueue_PriorityHigherFirst(t *testing.T) {
	q := NewQueue()
	base := time.Now()
	q.Push("low", 1, base)
	q.Push("high", 5, base)
	q.Push("mid", 3, base)

	want := []string{"high", "mid", "low"}
	for _, w := range want {
		if got := popOrFail(t, q); got != w {
			t.Fatalf("expected %q, got %q", w, got)
		}
	}
}

func TestQueue_FIFOWithinSamePriority(t *testing.T) {
	q := NewQueue()
	base := time.Now()
	q.Push("first", 1, base)
	q.Push("second", 1, base.Add(time.Millisecond))
	q.Push("third", 1, base.Add(2*time.Millisecond))

	want := []string{"first", "second", "third"}
	for _, w := range want {
		if got := popOrFail(t, q); got != w {
			t.Fatalf("expected %q, got %q", w, got)
		}
	}
}

func TestQueue_PopBlocksUntilPush(t *testing.T) {
	q := NewQueue()

	done := make(chan string, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		id, ok := q.Pop(ctx)
		if !ok {
			done <- ""
			return
		}
		done <- id
	}()

	time.Sleep(20 * time.Millisecond) // let the goroutine block in Pop

	select {
	case <-done:
		t.Fatal("Pop returned before any Push")
	default:
	}

	q.Push("a", 0, time.Now())

	select {
	case got := <-done:
		if got != "a" {
			t.Fatalf("expected \"a\", got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("Pop did not wake up after Push")
	}
}

func TestQueue_PopRespectsContext(t *testing.T) {
	q := NewQueue()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	id, ok := q.Pop(ctx)
	elapsed := time.Since(start)

	if ok {
		t.Fatalf("expected ok=false after context timeout, got id=%q", id)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Pop did not return promptly after ctx cancel: %v", elapsed)
	}
}

func TestQueue_CloseUnblocksPop(t *testing.T) {
	q := NewQueue()

	done := make(chan bool, 1)
	go func() {
		_, ok := q.Pop(context.Background())
		done <- ok
	}()
	time.Sleep(20 * time.Millisecond)

	q.Close()

	select {
	case ok := <-done:
		if ok {
			t.Fatal("expected ok=false after Close")
		}
	case <-time.After(time.Second):
		t.Fatal("Pop did not return after Close")
	}
}

func TestQueue_Len(t *testing.T) {
	q := NewQueue()
	if q.Len() != 0 {
		t.Fatalf("empty queue len=%d, want 0", q.Len())
	}
	now := time.Now()
	q.Push("a", 0, now)
	q.Push("b", 0, now.Add(time.Millisecond))
	if q.Len() != 2 {
		t.Fatalf("len=%d, want 2", q.Len())
	}
	popOrFail(t, q)
	if q.Len() != 1 {
		t.Fatalf("len=%d, want 1 after one pop", q.Len())
	}
}

// TestQueue_ConcurrentPopNoLossNoDup checks each pushed id is observed exactly once. Run with -race.
func TestQueue_ConcurrentPopNoLossNoDup(t *testing.T) {
	q := NewQueue()
	const N = 500
	const workers = 8

	go func() {
		base := time.Now()
		for i := range N {
			q.Push(idOf(i), i%5, base.Add(time.Duration(i)*time.Microsecond))
		}
	}()

	var (
		mu      sync.Mutex
		seen    = make(map[string]int)
		count   atomic.Int64
		stopCtx = context.Background()
	)

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for {
				if count.Load() == N {
					return
				}
				ctx, cancel := context.WithTimeout(stopCtx, 2*time.Second)
				id, ok := q.Pop(ctx)
				cancel()
				if !ok {
					return
				}
				mu.Lock()
				seen[id]++
				mu.Unlock()
				count.Add(1)
			}
		})
	}
	wg.Wait()

	if int(count.Load()) != N {
		t.Fatalf("received %d items, want %d", count.Load(), N)
	}
	if len(seen) != N {
		t.Fatalf("saw %d distinct ids, want %d", len(seen), N)
	}
	dups := []string{}
	for id, c := range seen {
		if c != 1 {
			dups = append(dups, id)
		}
	}
	sort.Strings(dups)
	if len(dups) > 0 {
		t.Fatalf("duplicates observed: %v", dups)
	}
}

// TestQueue_CloseWakesAllWaiters: every blocked Pop must return ok=false. Run with -race.
func TestQueue_CloseWakesAllWaiters(t *testing.T) {
	q := NewQueue()
	const workers = 8

	var wg sync.WaitGroup
	results := make(chan bool, workers)
	for range workers {
		wg.Go(func() {
			_, ok := q.Pop(context.Background())
			results <- ok
		})
	}
	time.Sleep(50 * time.Millisecond)
	q.Close()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not wake all waiters")
	}
	close(results)
	for ok := range results {
		if ok {
			t.Fatal("expected every Pop to return ok=false after Close")
		}
	}
}
