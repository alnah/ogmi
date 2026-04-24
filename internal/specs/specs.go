package specs

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type SourceRequest struct {
	FlagPath string
	EnvPath  string
}

type Source struct {
	FS          fs.FS
	Description string
}

//go:embed specs/** specs/*/.gitkeep specs/*/*/.gitkeep specs/*/*/*/.gitkeep specs/*/*/*/*/.gitkeep specs/*/*/*/*/*/.gitkeep specs/*/*/*/*/*/*/.gitkeep
var bundledSpecs embed.FS

func Bundle() fs.FS {
	return bundledSpecs
}

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
			return os.MkdirAll(target, 0o755)
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
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create export directory %s: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write exported spec %s: %w", target, err)
		}
		return nil
	})
}
