// Package viewer provides the embedded Dashforge viewer.
package viewer

import (
	"embed"
	"io/fs"
)

//go:embed index.html
var viewerFS embed.FS

// FS returns the embedded viewer filesystem.
func FS() fs.FS {
	return viewerFS
}
