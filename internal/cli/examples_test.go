package cli_test

import "testing"

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
