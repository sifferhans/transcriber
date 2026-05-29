package jobs

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStore_CreateAndGet(t *testing.T) {
	s := NewStore(0)
	in := Job{ID: "a", Path: "/x.wav", Status: StatusPending, Priority: 3}
	s.Create(in)

	got, ok := s.Get("a")
	if !ok {
		t.Fatal("expected job to exist")
	}
	if got.ID != "a" || got.Path != "/x.wav" || got.Status != StatusPending || got.Priority != 3 {
		t.Fatalf("unexpected job: %+v", got)
	}
}

func TestStore_GetReturnsCopy(t *testing.T) {
	s := NewStore(0)
	s.Create(Job{ID: "a", Status: StatusPending, Progress: 0})

	snap, _ := s.Get("a")
	snap.Progress = 99
	if snap.Progress != 99 {
		t.Fatal("local copy did not accept mutation")
	}

	again, _ := s.Get("a")
	if again.Progress != 0 {
		t.Fatalf("Get returned a reference, not a copy: progress=%d", again.Progress)
	}
}

func TestStore_GetMissing(t *testing.T) {
	s := NewStore(0)
	if _, ok := s.Get("nope"); ok {
		t.Fatal("expected missing job to return ok=false")
	}
}

func TestStore_UpdateApplies(t *testing.T) {
	s := NewStore(0)
	s.Create(Job{ID: "a", Status: StatusPending, Progress: 0})

	ok := s.Update("a", func(j *Job) {
		j.Status = StatusRunning
		j.Progress = 42
	})
	if !ok {
		t.Fatal("expected Update to return true")
	}
	got, _ := s.Get("a")
	if got.Status != StatusRunning || got.Progress != 42 {
		t.Fatalf("update did not persist: %+v", got)
	}
}

func TestStore_UpdateMissing(t *testing.T) {
	s := NewStore(0)
	called := false
	ok := s.Update("nope", func(j *Job) { called = true })
	if ok {
		t.Fatal("expected Update on missing job to return false")
	}
	if called {
		t.Fatal("update fn should not be called for missing job")
	}
}

func TestStore_ListSortedByCreatedAt(t *testing.T) {
	s := NewStore(0)
	now := time.Now()
	s.Create(Job{ID: "second", CreatedAt: now.Add(2 * time.Second)})
	s.Create(Job{ID: "first", CreatedAt: now.Add(1 * time.Second)})
	s.Create(Job{ID: "third", CreatedAt: now.Add(3 * time.Second)})

	got := s.List()
	if len(got) != 3 {
		t.Fatalf("want 3 jobs, got %d", len(got))
	}
	want := []string{"first", "second", "third"}
	for i, j := range got {
		if j.ID != want[i] {
			t.Fatalf("position %d: want %s, got %s", i, want[i], j.ID)
		}
	}
}

func TestStore_CancelInvokesFunc(t *testing.T) {
	s := NewStore(0)
	s.Create(Job{ID: "a", Status: StatusRunning})

	_, cancel := context.WithCancel(context.Background())
	var called atomic.Bool
	wrapped := func() {
		called.Store(true)
		cancel()
	}
	s.SetCancel("a", wrapped)

	if !s.Cancel("a") {
		t.Fatal("expected Cancel to return true")
	}
	if !called.Load() {
		t.Fatal("cancel fn was not invoked")
	}
}

func TestStore_CancelUnknown(t *testing.T) {
	s := NewStore(0)
	if s.Cancel("nope") {
		t.Fatal("expected Cancel on missing id to return false")
	}
}

func TestStore_CancelAfterClear(t *testing.T) {
	s := NewStore(0)
	s.Create(Job{ID: "a"})

	var called atomic.Bool
	s.SetCancel("a", func() { called.Store(true) })
	s.ClearCancel("a")

	if s.Cancel("a") {
		t.Fatal("expected Cancel to return false after ClearCancel")
	}
	if called.Load() {
		t.Fatal("cancel fn should not have been called")
	}
}

// TestStore_ConcurrentAccess is meant to be run with `go test -race`.
func TestStore_ConcurrentAccess(t *testing.T) {
	s := NewStore(0)
	const N = 50

	for i := range N {
		s.Create(Job{ID: idOf(i), Status: StatusPending})
	}

	var wg sync.WaitGroup
	for i := range N {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for k := range 100 {
				s.Update(idOf(i), func(j *Job) {
					j.Progress = k
				})
			}
		}(i)
	}
	for i := range N {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for range 100 {
				_, _ = s.Get(idOf(i))
				_ = s.List()
			}
		}(i)
	}
	wg.Wait()

	for i := range N {
		got, _ := s.Get(idOf(i))
		if got.Progress != 99 {
			t.Fatalf("job %d: expected final progress 99, got %d", i, got.Progress)
		}
	}
}

func TestStore_EvictTerminalKeepsActive(t *testing.T) {
	s := NewStore(2)
	base := time.Now()

	for i := range 5 {
		id := idOf(i)
		s.Create(Job{ID: id, Status: StatusPending, CreatedAt: base.Add(time.Duration(i) * time.Second)})
		s.Update(id, func(j *Job) {
			j.Status = StatusCompleted
			j.EndedAt = base.Add(time.Duration(i) * time.Second)
		})
	}

	s.Create(Job{ID: "running", Status: StatusRunning, CreatedAt: base.Add(10 * time.Second)})
	s.Create(Job{ID: "queued", Status: StatusPending, CreatedAt: base.Add(11 * time.Second)})

	got := s.List()
	if len(got) != 4 {
		t.Fatalf("expected 4 jobs (2 terminal + 2 active), got %d: %+v", len(got), got)
	}

	ids := map[string]bool{}
	for _, j := range got {
		ids[j.ID] = true
	}
	if !ids["running"] || !ids["queued"] {
		t.Fatalf("active jobs were evicted: %+v", ids)
	}
	if !ids[idOf(3)] || !ids[idOf(4)] {
		t.Fatalf("expected the two newest terminal jobs to survive, got %+v", ids)
	}
	if ids[idOf(0)] || ids[idOf(1)] || ids[idOf(2)] {
		t.Fatalf("expected oldest terminal jobs to be evicted, got %+v", ids)
	}
}

func TestStore_EvictMixedTerminalStatuses(t *testing.T) {
	s := NewStore(2)
	base := time.Now()
	mk := func(id, status string, secondsOffset int) {
		s.Create(Job{ID: id, Status: StatusPending, CreatedAt: base})
		s.Update(id, func(j *Job) {
			j.Status = status
			j.EndedAt = base.Add(time.Duration(secondsOffset) * time.Second)
		})
	}
	mk("oldfail", StatusFailed, 1)
	mk("oldcancel", StatusCanceled, 2)
	mk("recentdone", StatusCompleted, 3)
	mk("newestfail", StatusFailed, 4)

	got := s.List()
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs after eviction, got %d", len(got))
	}
	survived := map[string]bool{}
	for _, j := range got {
		survived[j.ID] = true
	}
	if !survived["recentdone"] || !survived["newestfail"] {
		t.Fatalf("wrong jobs survived: %+v", survived)
	}
}

func TestStore_EvictClearsCancel(t *testing.T) {
	s := NewStore(1)
	base := time.Now()

	s.Create(Job{ID: "old", Status: StatusRunning, CreatedAt: base})
	s.SetCancel("old", func() {})
	s.Update("old", func(j *Job) {
		j.Status = StatusCompleted
		j.EndedAt = base.Add(time.Second)
	})

	s.Create(Job{ID: "new", Status: StatusRunning, CreatedAt: base.Add(2 * time.Second)})
	s.Update("new", func(j *Job) {
		j.Status = StatusCompleted
		j.EndedAt = base.Add(3 * time.Second)
	})

	if _, ok := s.Get("old"); ok {
		t.Fatal("expected old job to be evicted")
	}
	s.mu.RLock()
	_, hasCancel := s.cancels["old"]
	s.mu.RUnlock()
	if hasCancel {
		t.Fatal("expected cancel entry to be cleared for evicted job")
	}
}

func TestStore_NoEvictionWhenLimitDisabled(t *testing.T) {
	s := NewStore(0)
	base := time.Now()
	for i := range 20 {
		id := idOf(i)
		s.Create(Job{ID: id, Status: StatusPending, CreatedAt: base})
		s.Update(id, func(j *Job) {
			j.Status = StatusCompleted
			j.EndedAt = base.Add(time.Duration(i) * time.Second)
		})
	}
	if got := len(s.List()); got != 20 {
		t.Fatalf("expected no eviction with limit=0, got %d jobs", got)
	}
}

func TestStore_CreateOrGet_NoKeyAlwaysInserts(t *testing.T) {
	s := NewStore(0)
	first, created := s.CreateOrGet(Job{ID: "a", Status: StatusPending})
	if !created || first.ID != "a" {
		t.Fatalf("first insert: created=%v id=%q", created, first.ID)
	}
	second, created := s.CreateOrGet(Job{ID: "b", Status: StatusPending})
	if !created || second.ID != "b" {
		t.Fatalf("second insert: created=%v id=%q", created, second.ID)
	}
}

func TestStore_CreateOrGet_SameKeyReturnsExisting(t *testing.T) {
	s := NewStore(0)
	first, created := s.CreateOrGet(Job{ID: "a", IdempotencyKey: "k1", Status: StatusPending})
	if !created {
		t.Fatal("first call should create")
	}

	second, created := s.CreateOrGet(Job{ID: "b", IdempotencyKey: "k1", Status: StatusPending})
	if created {
		t.Fatal("second call should be a hit, not a create")
	}
	if second.ID != first.ID {
		t.Fatalf("returned id: got %q, want %q", second.ID, first.ID)
	}
	if got := len(s.List()); got != 1 {
		t.Fatalf("store should still hold 1 job, got %d", got)
	}
}

func TestStore_CreateOrGet_AfterEvictionTreatsKeyAsFresh(t *testing.T) {
	s := NewStore(1)
	base := time.Now()

	_, _ = s.CreateOrGet(Job{ID: "old", IdempotencyKey: "k1", Status: StatusPending, CreatedAt: base})
	s.Update("old", func(j *Job) {
		j.Status = StatusCompleted
		j.EndedAt = base.Add(time.Second)
	})
	// Evict "old" by inserting another terminal job past the cap.
	_, _ = s.CreateOrGet(Job{ID: "new", Status: StatusPending, CreatedAt: base.Add(2 * time.Second)})
	s.Update("new", func(j *Job) {
		j.Status = StatusCompleted
		j.EndedAt = base.Add(3 * time.Second)
	})
	if _, ok := s.Get("old"); ok {
		t.Fatal("setup: expected 'old' to be evicted")
	}

	// Same key, new job ID — should insert since the previous mapping is gone.
	got, created := s.CreateOrGet(Job{ID: "fresh", IdempotencyKey: "k1", Status: StatusPending, CreatedAt: base.Add(4 * time.Second)})
	if !created {
		t.Fatal("expected re-use of key after eviction to create a new job")
	}
	if got.ID != "fresh" {
		t.Fatalf("id: got %q, want %q", got.ID, "fresh")
	}
}

func idOf(i int) string {
	return "id-" + itoa(i)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
