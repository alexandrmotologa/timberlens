package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

func getMimeType(p string) string {
	lower := strings.ToLower(p)
	switch {
	case strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".mjs"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(lower, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(lower, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(lower, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(lower, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".ico"):
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

// RegisterStaticHandlers attaches static file serving from the embedded React dist.
func (s *Server) RegisterStaticHandlers() {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return
	}

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If requesting an API route, do not serve static files
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		filePath := strings.TrimPrefix(r.URL.Path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Try opening the file in embedded FS
		data, err := fs.ReadFile(sub, filePath)
		if err == nil {
			w.Header().Set("Content-Type", getMimeType(filePath))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}

		// Fallback to index.html for SPA client-side routing
		indexBytes, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexBytes)
	})
}
