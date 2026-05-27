// Package hfcache resolves and caches Hugging Face model files on the local filesystem.
package hfcache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultBaseURL = "https://huggingface.co"

type Cache struct {
	root    string
	baseURL string
	client  *http.Client

	mu    sync.Mutex
	locks map[string]*sync.Mutex // serialize concurrent Get calls per local path
}

func New(root string) *Cache {
	return &Cache{
		root:    root,
		baseURL: defaultBaseURL,
		client:  &http.Client{}, // no timeout: model downloads can run for minutes
		locks:   map[string]*sync.Mutex{},
	}
}

// Default returns a Cache rooted at $XDG_CACHE_HOME/transcriber/hf (or ~/.cache).
func Default() *Cache {
	return New(filepath.Join(userCacheDir(), "transcriber", "hf"))
}

// WithBaseURL overrides the Hugging Face host (used in tests).
func (c *Cache) WithBaseURL(url string) *Cache {
	c.baseURL = url
	return c
}

// Get returns a local path for repo/file, downloading atomically on cache miss.
func (c *Cache) Get(ctx context.Context, repo, file string) (string, error) {
	if repo == "" || file == "" {
		return "", errors.New("hfcache: repo and file required")
	}
	local := filepath.Join(c.root, filepath.FromSlash(repo), file)

	lock := c.lockFor(local)
	lock.Lock()
	defer lock.Unlock()

	if _, err := os.Stat(local); err == nil {
		return local, nil
	}

	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/resolve/main/%s", c.baseURL, repo, file)
	slog.Info("hfcache: downloading", "repo", repo, "file", file, "to", local)
	start := time.Now()

	partial := local + ".partial"
	_ = os.Remove(partial)

	if err := c.download(ctx, url, partial); err != nil {
		_ = os.Remove(partial)
		return "", err
	}
	if err := os.Rename(partial, local); err != nil {
		return "", fmt.Errorf("hfcache rename: %w", err)
	}

	slog.Info("hfcache: downloaded", "repo", repo, "file", file, "took", time.Since(start))
	return local, nil
}

func (c *Cache) lockFor(path string) *sync.Mutex {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.locks[path]
	if !ok {
		m = &sync.Mutex{}
		c.locks[path] = m
	}
	return m
}

func (c *Cache) download(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("hfcache get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("hfcache get %s: %s", url, resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("hfcache write %s: %w", dest, err)
	}
	return f.Sync()
}

func userCacheDir() string {
	if v := os.Getenv("XDG_CACHE_HOME"); v != "" {
		return v
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".cache")
	}
	return os.TempDir()
}
