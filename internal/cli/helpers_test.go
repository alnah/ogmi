package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/cli"
)

type commandResult struct {
	exitCode int
	stdout   string
	stderr   string
}

type successEnvelope struct {
	Kind          string `json:"kind"`
	SchemaVersion string `json:"schemaVersion"`
}

type errorEnvelope struct {
	Kind          string `json:"kind"`
	SchemaVersion string `json:"schemaVersion"`
	Error         struct {
		Code        string       `json:"code"`
		Message     string       `json:"message"`
		Suggestions []string     `json:"suggestions"`
		Details     errorDetails `json:"details"`
	} `json:"error"`
}

type errorDetails struct {
	InvalidFilter   invalidFilter `json:"invalidFilter"`
	AvailableFields []string      `json:"availableFields"`
}

type invalidFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type descriptorListEnvelope struct {
	Kind          string                 `json:"kind"`
	SchemaVersion string                 `json:"schemaVersion"`
	Total         int                    `json:"total"`
	Items         []descriptorListRecord `json:"items"`
}

type descriptorListRecord struct {
	Corpus      string `json:"corpus"`
	Scale       string `json:"scale"`
	Level       string `json:"level"`
	Code        string `json:"code"`
	ID          string `json:"id"`
	Description string `json:"description"`
	File        string `json:"file"`
}

func runOgmi(t *testing.T, args ...string) commandResult {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Run(context.Background(), args, &stdout, &stderr)

	return commandResult{
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}
}

func requireSuccess(t *testing.T, result commandResult) {
	t.Helper()
	if result.exitCode != cli.ExitSuccess {
		t.Fatalf(
			"cli.Run() exit code = %d, want %d; stdout %q stderr %q",
			result.exitCode,
			cli.ExitSuccess,
			result.stdout,
			result.stderr,
		)
	}
	if result.stderr != "" {
		t.Errorf("cli.Run() stderr = %q, want empty", result.stderr)
	}
}

func requireJSONKind(t *testing.T, result commandResult, wantKind string) {
	t.Helper()
	requireSuccess(t, result)
	envelope := decodeSuccessEnvelope(t, result.stdout)
	want := successEnvelope{Kind: wantKind, SchemaVersion: cli.SchemaVersion}
	if diff := cmp.Diff(want, envelope); diff != "" {
		t.Errorf("JSON envelope mismatch (-want +got):\n%s", diff)
	}
}

func requireContainsAll(t *testing.T, text string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(text, value) {
			t.Errorf("output missing %q in:\n%s", value, text)
		}
	}
}

func requireCLIError(
	t *testing.T,
	result commandResult,
	wantExitCode int,
	wantCode string,
	wantMessageFragment string,
) errorEnvelope {
	t.Helper()
	if result.exitCode != wantExitCode {
		t.Errorf(
			"cli.Run() exit code = %d, want %d; stdout %q stderr %q",
			result.exitCode,
			wantExitCode,
			result.stdout,
			result.stderr,
		)
	}
	if result.stdout != "" {
		t.Errorf("cli.Run() stdout = %q, want empty on error", result.stdout)
	}

	var envelope errorEnvelope
	if err := json.Unmarshal([]byte(result.stderr), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v; stderr %q", err, result.stderr)
	}
	if envelope.Kind != "error" || envelope.SchemaVersion != cli.SchemaVersion || envelope.Error.Code != wantCode {
		t.Errorf("error envelope = %+v, want kind error schema %s code %s", envelope, cli.SchemaVersion, wantCode)
	}
	if !strings.Contains(envelope.Error.Message, wantMessageFragment) {
		t.Errorf("error message = %q, want it to mention %q", envelope.Error.Message, wantMessageFragment)
	}
	return envelope
}

func decodeSuccessEnvelope(t *testing.T, stdout string) successEnvelope {
	t.Helper()
	var envelope successEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v; stdout %q", err, stdout)
	}
	return envelope
}

func decodeDescriptorList(t *testing.T, stdout string) descriptorListEnvelope {
	t.Helper()
	var envelope descriptorListEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(descriptor list) error = %v; stdout %q", err, stdout)
	}
	return envelope
}

func writeThemeSpecsFixture(t *testing.T, idStem, description string) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, "specs", "themes")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(theme spec dir) error = %v", err)
	}
	yaml := strings.Join([]string{
		`id: "themes.descriptors"`,
		`version: "1"`,
		``,
		`catalog:`,
		`  - code: "agent_contract"`,
		`    id: "themes.descriptors.agent_contract"`,
		`    description:`,
		`      - "Agent contract fixtures."`,
		``,
		`entries:`,
		`  - scale: "agent_contract"`,
		`    level: "a1"`,
		`    code: "` + idStem + `"`,
		`    id: "themes.descriptors.agent_contract.` + idStem + `.a1"`,
		`    description: "` + description + `"`,
		``,
	}, "\n")
	if err := os.WriteFile(filepath.Join(specDir, "descriptors.yml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("os.WriteFile(theme descriptors.yml) error = %v", err)
	}
	return root
}
