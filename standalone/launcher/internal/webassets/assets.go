// Package webassets owns the embedded OHIF viewer dist.
//
// The `web/` subfolder is populated by standalone/scripts/build-viewer.sh and
// must sit next to this file because //go:embed paths are resolved relative
// to the Go source file that declares them.
package webassets

import (
	"embed"
	"io/fs"
)

//go:embed all:web
var embeddedFS embed.FS

// FS returns the embedded viewer dist rooted at "web/" so callers can treat
// it as if it were the dist directory itself.
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "web")
}
