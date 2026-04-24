package cli_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
