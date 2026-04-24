package specs_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/alnah/ogmi/internal/specs"
)

func TestResolveSourcePrecedence(t *testing.T) {
	tests := []struct {
		name       string
		request    specs.SourceRequest
		wantSource string
	}{
		{name: "flag wins", request: specs.SourceRequest{FlagPath: "/flag/specs", EnvPath: "/env/specs"}, wantSource: "/flag/specs"},
		{name: "env replaces embedded", request: specs.SourceRequest{EnvPath: "/env/specs"}, wantSource: "/env/specs"},
		{name: "embedded default", request: specs.SourceRequest{}, wantSource: "embedded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := specs.Resolve(tt.request)
			if err != nil {
				t.Fatalf("specs.Resolve() error = %v", err)
			}
			if got.Description != tt.wantSource {
				t.Errorf("specs.Resolve() source = %q, want %q", got.Description, tt.wantSource)
			}
		})
	}
}

func TestExportCopiesEmbeddedSpecsAndPreservesGitkeep(t *testing.T) {
	outputDir := t.TempDir()
	bundle := fstest.MapFS{
		"specs/cefr/production/speaking/descriptors.yml": {Data: []byte("entries: []\n")},
		"specs/cefr/production/speaking/.gitkeep":        {Data: []byte{}},
		"specs/themes/descriptors.yml":                   {Data: []byte("entries: []\n")},
		"specs/themes/.gitkeep":                          {Data: []byte{}},
	}

	if err := specs.Export(bundle, outputDir, false); err != nil {
		t.Fatalf("specs.Export() error = %v", err)
	}

	wantFiles := []string{
		"specs/cefr/production/speaking/descriptors.yml",
		"specs/cefr/production/speaking/.gitkeep",
		"specs/themes/descriptors.yml",
		"specs/themes/.gitkeep",
	}
	for _, name := range wantFiles {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("exported file %s stat error = %v", name, err)
		}
	}
}

func TestExportRefusesOverwriteWithoutForce(t *testing.T) {
	outputDir := t.TempDir()
	bundle := fstest.MapFS{"specs/themes/descriptors.yml": {Data: []byte("entries: []\n")}}
	if err := os.MkdirAll(filepath.Join(outputDir, "specs/themes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "specs/themes/descriptors.yml"), []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := specs.Export(bundle, outputDir, false); err == nil {
		t.Fatal("specs.Export(force=false) error = nil, want overwrite protection")
	}
	if err := specs.Export(bundle, outputDir, true); err != nil {
		t.Fatalf("specs.Export(force=true) error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(outputDir, "specs/themes/descriptors.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "entries: []\n" {
		t.Errorf("overwritten file = %q, want embedded content", got)
	}
}

func TestBundleContainsYMLDescriptorsAndGitkeepFiles(t *testing.T) {
	bundle := specs.Bundle()
	mustExist := []string{
		"specs/cefr/production/speaking/descriptors.yml",
		"specs/cefr/production/speaking/.gitkeep",
		"specs/french/grammar/descriptors.yml",
		"specs/texts/reading/descriptors.yml",
		"specs/themes/descriptors.yml",
		"specs/themes/.gitkeep",
	}
	for _, name := range mustExist {
		if _, err := fs.Stat(bundle, name); err != nil {
			t.Errorf("fs.Stat(bundle, %q) error = %v", name, err)
		}
	}
	if _, err := fs.Stat(bundle, "specs/themes/descriptors.yaml"); err == nil {
		t.Error("fs.Stat(bundle, descriptors.yaml) error = nil, want .yaml absent for v1")
	}
}
