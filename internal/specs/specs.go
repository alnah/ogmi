package specs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	ogmi "github.com/alnah/ogmi"
)

// SourceRequest declares explicit and environment spec source paths.
type SourceRequest struct {
	FlagPath string
	EnvPath  string
}

// Source contains the resolved specs filesystem and user-facing description.
type Source struct {
	FS          fs.FS
	Description string
}

// Bundle returns the embedded descriptor specs filesystem.
func Bundle() fs.FS {
	return ogmi.EmbeddedSpecs()
}

// Resolve selects specs from flag path, environment path, or embedded bundle.
func Resolve(request SourceRequest) (Source, error) {
	if request.FlagPath != "" {
		return Source{FS: os.DirFS(request.FlagPath), Description: request.FlagPath}, nil
	}
	envPath := request.EnvPath
	if envPath == "" {
		envPath = os.Getenv("OGMI_SPECS")
	}
	if envPath != "" {
		return Source{FS: os.DirFS(envPath), Description: envPath}, nil
	}
	return Source{FS: Bundle(), Description: "embedded"}, nil
}

// Export copies specs from fsys into outputDir, refusing overwrites unless forced.
func Export(fsys fs.FS, outputDir string, force bool) error {
	if outputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	return fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		target := filepath.Join(outputDir, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755) //nolint:gosec // Exported specs are user-traversable.
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("read embedded spec %s: %w", name, err)
		}
		if !force {
			if _, err := os.Stat(target); err == nil {
				return fmt.Errorf("refusing to overwrite %s without force", target)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("stat export target %s: %w", target, err)
			}
		}
		targetDir := filepath.Dir(target)
		if err := os.MkdirAll(targetDir, 0o755); err != nil { //nolint:gosec // Exported specs are user-traversable.
			return fmt.Errorf("create export directory %s: %w", targetDir, err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil { //nolint:gosec // Exported specs are user-readable.
			return fmt.Errorf("write exported spec %s: %w", target, err)
		}
		return nil
	})
}
