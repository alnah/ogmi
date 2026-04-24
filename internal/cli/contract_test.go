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
		t.Fatalf("cli.Run() exit code = %d, want %d; stdout %q stderr %q", result.exitCode, cli.ExitSuccess, result.stdout, result.stderr)
	}
	if result.stderr != "" {
		t.Errorf("cli.Run() stderr = %q, want empty", result.stderr)
	}
}

func decodeSuccessEnvelope(t *testing.T, stdout string) successEnvelope {
	t.Helper()
	var envelope successEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v; stdout %q", err, stdout)
	}
	return envelope
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

func decodeDescriptorList(t *testing.T, stdout string) descriptorListEnvelope {
	t.Helper()
	var envelope descriptorListEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(descriptor list) error = %v; stdout %q", err, stdout)
	}
	return envelope
}

func TestRootHelpDiscoversCommonCommands(t *testing.T) {
	result := runOgmi(t, "--help")
	requireSuccess(t, result)
	requireContainsAll(t, result.stdout,
		"Ogmi",
		"Common Commands",
		"descriptors",
		"specs",
		"version",
		"--format",
	)
}

func TestVersionCommandWritesStdoutOnly(t *testing.T) {
	result := runOgmi(t, "version")
	requireSuccess(t, result)
	requireContainsAll(t, strings.ToLower(result.stdout), "ogmi", "version")
}

func TestDescriptorsHelpDocumentsWorkflowAndExamples(t *testing.T) {
	result := runOgmi(t, "descriptors", "--help")
	requireSuccess(t, result)
	requireContainsAll(t, result.stdout,
		"inspect corpora",
		"inspect fields",
		"list descriptors",
		"get descriptor",
		"Examples:",
		"corpora",
		"fields",
		"schema",
		"list",
		"scales",
		"get",
		"compare-levels",
		"coverage",
		"examples",
	)
}

func TestEveryDescriptorCommandHasHelpExample(t *testing.T) {
	commands := []string{"corpora", "fields", "schema", "list", "scales", "get", "compare-levels", "coverage", "examples"}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			result := runOgmi(t, "descriptors", command, "--help")
			requireSuccess(t, result)
			requireContainsAll(t, result.stdout, "Usage:", "Examples:", "ogmi descriptors "+command)
		})
	}
}

func TestDescriptorCommandsDefaultToTypedJSON(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantKind string
	}{
		{name: "corpora", args: []string{"descriptors", "corpora"}, wantKind: "descriptor_corpora"},
		{name: "fields", args: []string{"descriptors", "fields"}, wantKind: "descriptor_fields"},
		{name: "schema", args: []string{"descriptors", "schema"}, wantKind: "descriptor_schema"},
		{name: "schema values", args: []string{"descriptors", "schema", "--field", "level", "--corpus", "cefr"}, wantKind: "descriptor_schema_values"},
		{name: "list", args: []string{"descriptors", "list", "--corpus", "cefr", "--level", "a1", "--limit", "1"}, wantKind: "descriptor_list"},
		{name: "scales", args: []string{"descriptors", "scales", "--corpus", "cefr", "--limit", "1"}, wantKind: "descriptor_scales"},
		{name: "get", args: []string{"descriptors", "get", "--corpus", "cefr", "--id", "cefr.production.speaking.descriptors.addressing_audiences.use_very_short_prepared_text_to_deliver_rehearsed_statement.a1"}, wantKind: "descriptor_get"},
		{name: "compare levels", args: []string{"descriptors", "compare-levels", "--corpus", "cefr", "--scale", "addressing_audiences", "--level", "a1", "--level", "b1"}, wantKind: "descriptor_level_comparison"},
		{name: "coverage", args: []string{"descriptors", "coverage", "--corpus", "cefr", "--domain", "production"}, wantKind: "descriptor_coverage_matrix"},
		{name: "examples", args: []string{"descriptors", "examples"}, wantKind: "descriptor_examples"},
		{name: "spec export", args: []string{"specs", "export", "--output", t.TempDir()}, wantKind: "specs_export"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runOgmi(t, tt.args...)
			requireJSONKind(t, result, tt.wantKind)
		})
	}
}

func TestTextFormatIsOptInForDataCommands(t *testing.T) {
	result := runOgmi(t, "--format", "text", "descriptors", "corpora")
	requireSuccess(t, result)
	if json.Valid([]byte(result.stdout)) {
		t.Fatalf("text output is valid JSON, want human-readable text: %q", result.stdout)
	}
	requireContainsAll(t, result.stdout, "cefr", "french", "texts", "themes")
}

func TestErrorsUseStderrAndStableExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantCode     string
		wantMessage  string
	}{
		{name: "domain", args: []string{"descriptors", "list", "--corpus", "cefrr"}, wantExitCode: cli.ExitDomain, wantCode: "unknown_corpus", wantMessage: "cefrr"},
		{name: "usage", args: []string{"descriptors", "get", "--corpus", "cefr"}, wantExitCode: cli.ExitUsage, wantCode: "usage", wantMessage: "--id"},
		{name: "invalid flag", args: []string{"descriptors", "list", "--sort-order", "sideways"}, wantExitCode: cli.ExitUsage, wantCode: "usage", wantMessage: "sort-order"},
		{name: "missing external specs", args: []string{"--specs", filepath.Join(t.TempDir(), "missing"), "descriptors", "corpora"}, wantExitCode: cli.ExitDomain, wantCode: "missing_specs", wantMessage: "missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runOgmi(t, tt.args...)
			if result.exitCode != tt.wantExitCode {
				t.Errorf("cli.Run(%q) exit code = %d, want %d", tt.args, result.exitCode, tt.wantExitCode)
			}
			if result.stdout != "" {
				t.Errorf("cli.Run(%q) stdout = %q, want empty on error", tt.args, result.stdout)
			}

			var envelope errorEnvelope
			if err := json.Unmarshal([]byte(result.stderr), &envelope); err != nil {
				t.Fatalf("json.Unmarshal(stderr) error = %v; stderr %q", err, result.stderr)
			}
			if envelope.Kind != "error" || envelope.SchemaVersion != cli.SchemaVersion || envelope.Error.Code != tt.wantCode {
				t.Errorf("error envelope = %+v, want kind error schema %s code %s", envelope, cli.SchemaVersion, tt.wantCode)
			}
			if !strings.Contains(envelope.Error.Message, tt.wantMessage) {
				t.Errorf("error message = %q, want it to mention %q", envelope.Error.Message, tt.wantMessage)
			}
		})
	}
}

func TestCLIReportsTypedStructuredFieldErrors(t *testing.T) {
	result := runOgmi(t, "descriptors", "schema", "--field", "subdomian")
	if result.exitCode != cli.ExitDomain {
		t.Fatalf("cli.Run(schema unknown field) exit code = %d, want %d; stdout %q stderr %q", result.exitCode, cli.ExitDomain, result.stdout, result.stderr)
	}

	var envelope errorEnvelope
	if err := json.Unmarshal([]byte(result.stderr), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v; stderr %q", err, result.stderr)
	}
	if envelope.Error.Code != "unknown_field" {
		t.Errorf("error code = %q, want unknown_field", envelope.Error.Code)
	}
	wantInvalidFilter := invalidFilter{Field: "field", Value: "subdomian"}
	if diff := cmp.Diff(wantInvalidFilter, envelope.Error.Details.InvalidFilter); diff != "" {
		t.Errorf("invalidFilter mismatch (-want +got):\n%s", diff)
	}
	wantAvailableFields := []string{"corpus", "domain", "subdomain", "scale", "level", "code", "id"}
	if diff := cmp.Diff(wantAvailableFields, envelope.Error.Details.AvailableFields); diff != "" {
		t.Errorf("availableFields mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"subdomain"}, envelope.Error.Suggestions); diff != "" {
		t.Errorf("suggestions mismatch (-want +got):\n%s", diff)
	}
}

func TestSpecsFlagChangesDescriptorData(t *testing.T) {
	fixtureRoot := writeThemeSpecsFixture(t, "from_flag", "Loaded from --specs.")
	result := runOgmi(t, "--specs", fixtureRoot, "descriptors", "list", "--corpus", "themes")
	requireJSONKind(t, result, "descriptor_list")

	list := decodeDescriptorList(t, result.stdout)
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("descriptor list total/items = %d/%d, want 1/1; stdout %s", list.Total, len(list.Items), result.stdout)
	}
	want := descriptorListRecord{Corpus: "themes", Scale: "agent_contract", Level: "a1", Code: "from_flag", ID: "themes.descriptors.agent_contract.from_flag.a1", Description: "Loaded from --specs.", File: "specs/themes/descriptors.yml"}
	if diff := cmp.Diff(want, list.Items[0]); diff != "" {
		t.Errorf("descriptor loaded through --specs mismatch (-want +got):\n%s", diff)
	}
}

func TestOGMISpecsChangesDescriptorDataWhenFlagAbsent(t *testing.T) {
	fixtureRoot := writeThemeSpecsFixture(t, "from_env", "Loaded from OGMI_SPECS.")
	t.Setenv("OGMI_SPECS", fixtureRoot)

	result := runOgmi(t, "descriptors", "list", "--corpus", "themes")
	requireJSONKind(t, result, "descriptor_list")
	list := decodeDescriptorList(t, result.stdout)
	if got, want := list.Items[0].ID, "themes.descriptors.agent_contract.from_env.a1"; got != want {
		t.Errorf("descriptor id with OGMI_SPECS = %q, want %q", got, want)
	}
}

func TestSpecsFlagWinsOverOGMISpecs(t *testing.T) {
	envRoot := writeThemeSpecsFixture(t, "from_env", "Loaded from OGMI_SPECS.")
	flagRoot := writeThemeSpecsFixture(t, "from_flag", "Loaded from --specs.")
	t.Setenv("OGMI_SPECS", envRoot)

	result := runOgmi(t, "--specs", flagRoot, "descriptors", "list", "--corpus", "themes")
	requireJSONKind(t, result, "descriptor_list")
	list := decodeDescriptorList(t, result.stdout)
	if got, want := list.Items[0].ID, "themes.descriptors.agent_contract.from_flag.a1"; got != want {
		t.Errorf("descriptor id with --specs and OGMI_SPECS = %q, want flag source %q", got, want)
	}
}

func TestDescriptorExamplesReturnAgentWorkflows(t *testing.T) {
	result := runOgmi(t, "descriptors", "examples")
	requireJSONKind(t, result, "descriptor_examples")
	requireContainsAll(t, result.stdout,
		"descriptors corpora",
		"descriptors list --corpus cefr --domain production --subdomain speaking --level a1",
		"descriptors get --corpus cefr --id",
		"descriptors scales --corpus cefr",
		"descriptors schema --field level",
		"descriptors coverage --corpus cefr",
		"specs export",
		"--specs",
	)
}
