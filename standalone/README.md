# OHIF Standalone Viewer

A zero-install, single-file binary that ships the OHIF Viewer plus a small Go
web server. Drop DICOM files into a `study/` folder next to the binary,
double-click it, and the viewer opens in the default browser.

```
dist/
├── macos_view                 (universal if built on macOS with lipo)
├── linux_view
├── windows_view.exe
└── study/
    ├── IM0001.dcm
    ├── IM0002.dcm
    └── ...
```

The binary:

1. Scans `./study/` for DICOM files (any extension, recursive).
2. Builds a DICOM JSON Model manifest in memory.
3. Serves the embedded OHIF viewer at `http://127.0.0.1:<port>/`.
4. Opens the browser at `?datasources=dicomjson&url=/study.json&StudyInstanceUIDs=<uid>`.

OHIF reads the manifest, fetches DICOM instances from `/study/...` through
the cornerstone DICOM image loader, and renders them.

## Layout

```
standalone/
├── viewer-config/
│   └── standalone.js          OHIF runtime config (dicomjson data source)
├── launcher/                  Go module (go.mod)
│   ├── cmd/viewer/main.go     Entry point: embed, scan, serve, open
│   ├── internal/
│   │   ├── browser/           OS-specific "open URL"
│   │   ├── dicom/             DICOM scan + DICOM JSON Model builder
│   │   ├── server/            Static + manifest + /study/* HTTP routing
│   │   └── webassets/         //go:embed of the OHIF dist (auto-populated)
├── scripts/
│   ├── build-viewer.sh        Builds OHIF with APP_CONFIG=config/standalone.js
│   └── package.sh             Cross-compiles macos/linux/windows binaries
├── Makefile
├── README.md
└── .gitignore
```

## Build

Requirements:

- Node 18+, Yarn 1.22+, Go 1.22+
- (macOS only, optional) `lipo` for a universal binary

From the repository root:

```bash
# 1. Install JS dependencies (once)
yarn install --frozen-lockfile

# 2. Build OHIF + cross-compile the launchers
make -C standalone all
```

Outputs land in `standalone/dist/`:

```
standalone/dist/
├── macos_view          (universal on macOS, arm64 otherwise)
├── linux_view
├── windows_view.exe
└── study/
    └── README.txt
```

### Individual targets

```bash
make -C standalone viewer     # just rebuild the OHIF dist + copy to launcher/web
make -C standalone launcher   # just build a dev binary (current OS)
make -C standalone package    # only cross-compile (needs viewer output)
make -C standalone run        # dev run against standalone/example-study/
make -C standalone clean
```

## Runtime flags

```
macos_view --help
  --study string   path to STUDY folder (default: ./study next to binary)
  --web    string  override viewer dist dir (defaults to embedded)
  --addr   string  listen address (default 127.0.0.1:0, i.e. random free port)
  --open   bool    open default browser automatically (default true)
```

Examples:

```bash
# Serve a specific folder, don't open the browser
./macos_view --study /path/to/study --open=false

# Hot-reload viewer assets from disk during development
./macos_view --web ./launcher/internal/webassets/web
```

## Architecture notes

- Single data source: `dicomjson` (no server-side DICOMweb required).
- Instance URLs inside the manifest are prefixed with `dicomweb:` and point
  at `/study/<relative path>`; cornerstone's WADO-URI loader fetches them
  as raw DICOM P10 files.
- All files are served from `127.0.0.1` only. No external network access.
- `Cross-Origin-Opener-Policy: same-origin` and `Cross-Origin-Embedder-Policy:
  require-corp` headers are set so cornerstone can use SharedArrayBuffer.

## Limits / to-do

- Single study at a time (the manifest groups by `StudyInstanceUID` but the
  launched URL opens the first one only).
- No lazy loading at the DICOMweb level; every file in `study/` is inspected
  once at startup. For huge studies consider switching to a mini QIDO/WADO-RS
  server (server/dicomweb.go is a placeholder).
- DICOMDIR files are ignored; the scanner only reads individual `.dcm` parts.
