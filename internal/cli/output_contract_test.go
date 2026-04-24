package cli_test

import (
	"encoding/json"
	"testing"
)

const knownDescriptorID = "cefr.production.speaking.descriptors.addressing_audiences.use_very_short_prepared_text_to_deliver_rehearsed_statement.a1"

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
		{name: "list with field flags", args: []string{"descriptors", "list", "--corpus", "cefr", "--level", "a1", "--limit", "1", "--sort-by", "code", "--group-by", "level"}, wantKind: "descriptor_list"},
		{name: "scales", args: []string{"descriptors", "scales", "--corpus", "cefr", "--limit", "1"}, wantKind: "descriptor_scales"},
		{name: "scales with sort field", args: []string{"descriptors", "scales", "--corpus", "cefr", "--limit", "1", "--sort-by", "code"}, wantKind: "descriptor_scales"},
		{name: "get", args: []string{"descriptors", "get", "--corpus", "cefr", "--id", knownDescriptorID}, wantKind: "descriptor_get"},
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

func TestSelectedDataCommandsWriteHumanReadableText(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{name: "list", args: []string{"--format", "text", "descriptors", "list", "--corpus", "cefr", "--limit", "1"}, contains: []string{"Descriptors:"}},
		{name: "scales", args: []string{"--format", "text", "descriptors", "scales", "--corpus", "cefr", "--limit", "1"}, contains: []string{"Descriptor scales:"}},
		{name: "get", args: []string{"--format", "text", "descriptors", "get", "--corpus", "cefr", "--id", knownDescriptorID}, contains: []string{"Descriptor:", knownDescriptorID}},
		{name: "compare levels", args: []string{"--format", "text", "descriptors", "compare-levels", "--corpus", "cefr", "--scale", "addressing_audiences", "--level", "a1"}, contains: []string{"Compared levels:"}},
		{name: "coverage", args: []string{"--format", "text", "descriptors", "coverage", "--corpus", "cefr", "--domain", "production", "--scale", "addressing_audiences"}, contains: []string{"Coverage total:"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runOgmi(t, tt.args...)
			requireSuccess(t, result)
			for _, value := range tt.contains {
				requireContainsAll(t, result.stdout, value)
			}
			if json.Valid([]byte(result.stdout)) {
				t.Fatalf("text output is valid JSON, want command summary: %q", result.stdout)
			}
		})
	}
}
