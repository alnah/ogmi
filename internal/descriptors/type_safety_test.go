package descriptors_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDescriptorPackagesAvoidGenericAnyDetails(t *testing.T) {
	packageRoots := []string{
		filepath.Join("..", "cli"),
		".",
		filepath.Join("..", "specs"),
	}

	var offenders []string
	for _, root := range packageRoots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(data)
			for _, forbidden := range []string{"map[string]any", "map[string]interface{}"} {
				if strings.Contains(text, forbidden) {
					offenders = append(offenders, path+": "+forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan production files under %s: %v", root, err)
		}
	}
	if len(offenders) > 0 {
		t.Fatalf("descriptor packages use generic detail blobs; introduce typed result/error details instead: %v", offenders)
	}
}
