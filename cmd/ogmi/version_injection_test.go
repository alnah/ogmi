package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReleaseBuildInjectsVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := filepath.Join(t.TempDir(), "ogmi")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	//nolint:gosec // Test builds the local cmd/ogmi package with controlled ldflags.
	build := exec.CommandContext(ctx, "go", "build", "-o", binaryPath,
		"-ldflags", "-X github.com/alnah/ogmi/internal/cli.version=0.1.0", ".")
	buildOutput, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build with release version ldflags error = %v; combined output:\n%s", err, buildOutput)
	}

	//nolint:gosec // Test executes a temporary binary built from the local cmd/ogmi package.
	version := exec.CommandContext(ctx, binaryPath, "version")
	versionOutput, err := version.CombinedOutput()
	if err != nil {
		t.Fatalf("release binary version command error = %v; combined output:\n%s", err, versionOutput)
	}
	if ctx.Err() != nil {
		t.Fatalf("release version injection test timed out: %v", ctx.Err())
	}
	if got, want := string(versionOutput), "ogmi version 0.1.0\n"; got != want {
		t.Fatalf("release binary version output = %q, want %q", got, want)
	}
}
