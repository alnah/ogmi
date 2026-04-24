package ogmi

import (
	"embed"
	"io/fs"
)

//go:embed all:specs
var embeddedSpecs embed.FS

// EmbeddedSpecs returns the bundled descriptor specs with paths rooted at specs/.
func EmbeddedSpecs() fs.FS {
	return embeddedSpecs
}
