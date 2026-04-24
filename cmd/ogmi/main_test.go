package main

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestEntrypointWiresCLIExitCodeAndStreams(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . --help error = %v; combined output:\n%s", err, output)
	}
	text := string(output)
	for _, want := range []string{"Ogmi", "descriptors", "specs", "version"} {
		if !strings.Contains(text, want) {
			t.Errorf("go run . --help output missing %q in:\n%s", want, text)
		}
	}
}
