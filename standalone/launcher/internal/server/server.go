package server

import (
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/bluepacs/standalone-viewer/internal/dicom"
)

// Options configures the standalone HTTP server.
type Options struct {
	// WebFS serves the OHIF viewer dist (index.html, assets, ...).
	WebFS fs.FS
	// StudyDir is the absolute path to the user's STUDY folder on disk.
	StudyDir string
	// Manifest is the pre-built DICOM JSON Model served at StudyURL.
	Manifest *dicom.Manifest
	// StudyURL is the path the manifest is served at. Default: /study.json
	StudyURL string
	// StudyPath is the URL prefix the raw DICOM files are served under.
	// It must match the `url` values generated inside the manifest.
	// Default: /study/
	StudyPath string
}

// New builds the HTTP handler that powers the standalone viewer.
func New(opts Options) http.Handler {
	if opts.StudyURL == "" {
		opts.StudyURL = "/study.json"
	}
	if opts.StudyPath == "" {
		opts.StudyPath = "/study/"
	}
	if !strings.HasSuffix(opts.StudyPath, "/") {
		opts.StudyPath += "/"
	}

	mux := http.NewServeMux()

	mux.HandleFunc(opts.StudyURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(opts.Manifest); err != nil {
			log.Printf("manifest encode error: %v", err)
		}
	})

	mux.Handle(opts.StudyPath, http.StripPrefix(strings.TrimRight(opts.StudyPath, "/"),
		dicomFileServer(http.Dir(opts.StudyDir))))

	webHandler := staticHandler(opts.WebFS)
	mux.Handle("/", webHandler)

	return withCommonHeaders(withLogger(mux))
}

// dicomFileServer wraps http.FileServer so that DICOM files are served with a
// conservative content-type and cache policy.
func dicomFileServer(root http.FileSystem) http.Handler {
	fs := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(r.URL.Path) == "" {
			w.Header().Set("Content-Type", "application/dicom")
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fs.ServeHTTP(w, r)
	})
}

// staticHandler serves the embedded/on-disk viewer dist with SPA-friendly
// fallback: if a non-asset path is requested and not found, stream index.html
// inline (without letting http.FileServer rewrite the URL to "./").
func staticHandler(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			serveIndex(w, root)
			return
		}
		if _, err := fs.Stat(root, clean); err != nil {
			if !hasAssetExt(clean) {
				serveIndex(w, root)
				return
			}
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

// serveIndex writes web/index.html inline. We avoid http.FileServer for this
// case because it issues a 301 for "/index.html" -> "./", which rewrites the
// client URL and breaks React-Router paths like /viewer/dicomjson.
func serveIndex(w http.ResponseWriter, root fs.FS) {
	f, err := root.Open("index.html")
	if err != nil {
		http.Error(w, "index.html missing", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, f); err != nil {
		log.Printf("index.html stream error: %v", err)
	}
}

func hasAssetExt(p string) bool {
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".js", ".mjs", ".css", ".map", ".json", ".wasm",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp",
		".woff", ".woff2", ".ttf", ".otf", ".eot",
		".html", ".txt", ".xml":
		return true
	}
	return false
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Required for SharedArrayBuffer / multi-threaded cornerstone decoders.
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
