package api

import (
	_ "embed"
	"net/http"
)

//go:embed docs/openapi.yaml
var openAPISpec []byte

//go:embed docs/index.html
var docsHTML []byte

// Redoc 2.5.3 — https://cdn.jsdelivr.net/npm/redoc@2.5.3/bundles/redoc.standalone.js
// Vendored so /docs doesn't depend on an external CDN and the script can't be
// swapped out from under us. Bump by re-downloading from the same path with a
// new version.
//
//go:embed docs/redoc.standalone.js
var redocJS []byte

func (s *Server) openapi(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(openAPISpec)
}

func (s *Server) docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(docsHTML)
}

func (s *Server) docsRedocJS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(redocJS)
}
