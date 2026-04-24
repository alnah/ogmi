package specs

import (
	"errors"
	"io/fs"
)

type SourceRequest struct {
	FlagPath string
	EnvPath  string
}

type Source struct {
	FS          fs.FS
	Description string
}

type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func Bundle() fs.FS {
	return emptyFS{}
}

func Resolve(request SourceRequest) (Source, error) {
	_ = request
	return Source{FS: Bundle(), Description: "embedded"}, nil
}

func Export(fsys fs.FS, outputDir string, force bool) error {
	_ = fsys
	_ = outputDir
	_ = force
	return errors.New("not implemented")
}
