// Package builder provides the embedded Dashforge dashboard builder.
package builder

import (
	"embed"
	"io/fs"
)

//go:embed dist
var builderFS embed.FS

// FS returns the embedded builder filesystem, rooted at dist/.
func FS() fs.FS {
	subFS, err := fs.Sub(builderFS, "dist")
	if err != nil {
		// This should never happen since dist is embedded
		panic("builder: failed to access dist directory: " + err.Error())
	}
	return subFS
}
