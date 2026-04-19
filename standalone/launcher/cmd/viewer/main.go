package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bluepacs/standalone-viewer/internal/browser"
	"github.com/bluepacs/standalone-viewer/internal/dicom"
	"github.com/bluepacs/standalone-viewer/internal/server"
	"github.com/bluepacs/standalone-viewer/internal/webassets"
)

func main() {
	var (
		studyDir   string
		webDir     string
		addr       string
		openBrowse bool
	)

	defaultStudyDir := defaultRelative("study")
	flag.StringVar(&studyDir, "study", defaultStudyDir, "path to the STUDY folder containing DICOM files")
	flag.StringVar(&webDir, "web", "", "override path to the OHIF viewer dist (defaults to embedded assets)")
	flag.StringVar(&addr, "addr", "127.0.0.1:0", "HTTP listen address (use 127.0.0.1:0 for a free port)")
	flag.BoolVar(&openBrowse, "open", true, "open the default browser automatically")
	flag.Parse()

	absStudyDir, err := filepath.Abs(studyDir)
	if err != nil {
		log.Fatalf("invalid --study path: %v", err)
	}
	if _, err := os.Stat(absStudyDir); os.IsNotExist(err) {
		log.Fatalf("study folder does not exist: %s", absStudyDir)
	}

	log.Printf("scanning study folder: %s", absStudyDir)
	start := time.Now()
	manifest, err := dicom.BuildManifest(absStudyDir)
	if err != nil {
		log.Fatalf("failed to scan study: %v", err)
	}
	log.Printf("scan complete: %d study, %d series, %d instances (%s)",
		len(manifest.Studies), manifest.CountSeries(), manifest.CountInstances(), time.Since(start))

	if len(manifest.Studies) == 0 {
		log.Fatalf("no DICOM instances found in %s", absStudyDir)
	}

	var webFS fs.FS
	if webDir != "" {
		info, err := os.Stat(webDir)
		if err != nil || !info.IsDir() {
			log.Fatalf("invalid --web path: %s", webDir)
		}
		webFS = os.DirFS(webDir)
		log.Printf("serving viewer assets from disk: %s", webDir)
	} else {
		sub, err := webassets.FS()
		if err != nil {
			log.Fatalf("failed to access embedded web assets: %v", err)
		}
		webFS = sub
		log.Printf("serving viewer assets from embedded bundle")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}
	tcpAddr := listener.Addr().(*net.TCPAddr)
	baseURL := fmt.Sprintf("http://%s:%d", tcpAddr.IP.String(), tcpAddr.Port)

	handler := server.New(server.Options{
		WebFS:     webFS,
		StudyDir:  absStudyDir,
		Manifest:  manifest,
		StudyURL:  "/study.json",
		StudyPath: "/study/",
	})

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	viewerURL := fmt.Sprintf("%s/viewer/dicomjson?url=/study.json&StudyInstanceUIDs=%s",
		baseURL, manifest.Studies[0].StudyInstanceUID)

	log.Printf("viewer ready: %s", viewerURL)

	if openBrowse {
		go func() {
			time.Sleep(200 * time.Millisecond)
			if err := browser.Open(viewerURL); err != nil {
				log.Printf("could not open browser: %v", err)
			}
		}()
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Printf("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// defaultRelative resolves a path relative to the executable location so that
// double-clicking the binary from its shipped folder "just works".
func defaultRelative(name string) string {
	exe, err := os.Executable()
	if err != nil {
		return name
	}
	return filepath.Join(filepath.Dir(exe), name)
}
