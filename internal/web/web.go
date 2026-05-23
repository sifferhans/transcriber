// Package web serves the embedded SPA bundle.
//
// The dist/ directory is populated by `make frontend` (which runs
// `pnpm generate` in frontend/ and copies the output here) and embedded
// into the binary at compile time. The `all:` prefix is required because
// Nuxt outputs assets under `_nuxt/` — directories starting with `_` are
// skipped by a plain `//go:embed`.
//
// During development, run the Nuxt dev server directly — Go ships the
// stale embedded copy and won't pick up frontend edits without a rebuild.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// Handler serves the embedded SPA: existing files are served as-is,
// anything else falls back to index.html so the SPA's client router can
// handle deep links (e.g. /jobs/{id}).
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("web: fs.Sub: " + err.Error())
	}
	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic("web: dist/index.html missing — run `make frontend`: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(sub, p); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Unknown path → SPA shell. Client router takes over.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}
