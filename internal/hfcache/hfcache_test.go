package hfcache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGetDownloadsAndCaches(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path != "/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("model bytes"))
	}))
	defer srv.Close()

	c := New(t.TempDir()).WithBaseURL(srv.URL)

	path, err := c.Get(context.Background(), "ggerganov/whisper.cpp", "ggml-tiny.bin")
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached: %v", err)
	}
	if string(got) != "model bytes" {
		t.Fatalf("contents = %q, want %q", got, "model bytes")
	}

	// Second call must hit the cache, not the server.
	path2, err := c.Get(context.Background(), "ggerganov/whisper.cpp", "ggml-tiny.bin")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if path2 != path {
		t.Errorf("cached path changed: %q vs %q", path, path2)
	}
	if h := atomic.LoadInt32(&hits); h != 1 {
		t.Errorf("server hits = %d, want 1 (second call should have hit cache)", h)
	}
}

func TestGet404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := New(t.TempDir()).WithBaseURL(srv.URL)
	_, err := c.Get(context.Background(), "ggerganov/whisper.cpp", "does-not-exist.bin")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	// Partial file should not be left behind.
	partial := filepath.Join(c.root, "ggerganov", "whisper.cpp", "does-not-exist.bin.partial")
	if _, err := os.Stat(partial); !os.IsNotExist(err) {
		t.Errorf(".partial leaked after failed download: %v", err)
	}
}

func TestConcurrentGetCoalesces(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("model"))
	}))
	defer srv.Close()

	c := New(t.TempDir()).WithBaseURL(srv.URL)

	const N = 20
	var wg sync.WaitGroup
	wg.Add(N)
	for range N {
		go func() {
			defer wg.Done()
			if _, err := c.Get(context.Background(), "repo", "file.bin"); err != nil {
				t.Errorf("Get: %v", err)
			}
		}()
	}
	wg.Wait()

	if h := atomic.LoadInt32(&hits); h != 1 {
		t.Errorf("server hits = %d, want 1 (concurrent Gets should coalesce)", h)
	}
}

func TestGetCleansStalePartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fresh"))
	}))
	defer srv.Close()

	c := New(t.TempDir()).WithBaseURL(srv.URL)
	// Simulate a leftover .partial from a previous crashed run.
	dir := filepath.Join(c.root, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dir, "file.bin.partial")
	if err := os.WriteFile(stale, []byte("stale junk"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := c.Get(context.Background(), "repo", "file.bin")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "fresh" {
		t.Errorf("contents = %q, want fresh download", got)
	}
}
