package cli_test

import (
	"strings"
	"testing"
)

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
