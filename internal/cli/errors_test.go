package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/cli"
)

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
		{name: "invalid format", args: []string{"--format", "yaml", "descriptors", "corpora"}, wantExitCode: cli.ExitUsage, wantCode: "usage", wantMessage: "--format"},
		{name: "missing specs export output", args: []string{"specs", "export"}, wantExitCode: cli.ExitUsage, wantCode: "usage", wantMessage: "--output"},
		{name: "missing external specs", args: []string{"--specs", filepath.Join(t.TempDir(), "missing"), "descriptors", "corpora"}, wantExitCode: cli.ExitDomain, wantCode: "missing_specs", wantMessage: "missing"},
		{name: "unknown root command", args: []string{"nope"}, wantExitCode: cli.ExitUsage, wantCode: "usage", wantMessage: "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runOgmi(t, tt.args...)
			_ = requireCLIError(t, result, tt.wantExitCode, tt.wantCode, tt.wantMessage)
		})
	}
}

func TestCLIReportsUsageErrorsForInvalidDescriptorCommandArguments(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		messageFragments []string
	}{
		{
			name:             "nested unknown command",
			args:             []string{"descriptors", "nope"},
			messageFragments: []string{"unknown command", "nope"},
		},
		{
			name:             "invalid scales sort order",
			args:             []string{"descriptors", "scales", "--corpus", "cefr", "--sort-order", "sideways", "--limit", "1"},
			messageFragments: []string{"--sort-order", "asc", "desc"},
		},
		{
			name:             "invalid list sort field",
			args:             []string{"descriptors", "list", "--corpus", "cefr", "--sort-by", "nope", "--limit", "1"},
			messageFragments: []string{"--sort-by", "nope"},
		},
		{
			name:             "invalid scales sort field",
			args:             []string{"descriptors", "scales", "--corpus", "cefr", "--sort-by", "nope", "--limit", "1"},
			messageFragments: []string{"--sort-by", "nope"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runOgmi(t, tt.args...)
			envelope := requireCLIError(t, result, cli.ExitUsage, "usage", tt.messageFragments[0])
			requireContainsAll(t, envelope.Error.Message, tt.messageFragments...)
		})
	}
}

func TestCLIReportsTypedStructuredFieldErrors(t *testing.T) {
	result := runOgmi(t, "descriptors", "schema", "--field", "subdomian")
	envelope := requireCLIError(t, result, cli.ExitDomain, "unknown_field", "subdomian")

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
