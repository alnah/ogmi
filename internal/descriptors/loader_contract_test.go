package descriptors_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
	"github.com/alnah/ogmi/internal/specs"
)

const cefrSpeakingYAML = `id: "cefr.production.speaking.descriptors"
version: "1"

catalog:
  - code: "addressing_audiences"
    id: "cefr.production.speaking.descriptors.addressing_audiences"
    description:
      - "Address audiences."
  - code: "turntaking"
    id: "cefr.production.speaking.descriptors.turntaking"
    description: "Manage turntaking."

entries:
  - scale: "addressing_audiences"
    level: "a2"
    code: "present_simple_announcement"
    id: "cefr.production.speaking.descriptors.addressing_audiences.present_simple_announcement.a2"
    description: "Can present a simple announcement."
  - scale: "addressing_audiences"
    level: "a1"
    code: "deliver_toast"
    id: "cefr.production.speaking.descriptors.addressing_audiences.deliver_toast.a1"
    description: "Can deliver a short toast."
  - scale: "turntaking"
    level: "b1"
    code: "maintain_exchange"
    id: "cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1"
    description:
      - "Can maintain an exchange."
      - "Can add relevant details."
`

const themesYAML = `id: "themes.descriptors"
version: "1"

catalog:
  - code: "personal_identity"
    id: "themes.descriptors.personal_identity"
    description:
      - "Personal identity themes."

entries:
  - scale: "personal_identity"
    level: "pre_a1"
    code: "personal_details"
    id: "themes.descriptors.personal_identity.personal_details.pre_a1"
    description: "Can share basic personal details."
`

func descriptorFixtureFS() fs.FS {
	return fstest.MapFS{
		"specs/cefr/production/speaking/descriptors.yml":  {Data: []byte(cefrSpeakingYAML)},
		"specs/cefr/production/speaking/descriptors.yaml": {Data: []byte("entries: []\n")},
		"specs/cefr/production/speaking/.gitkeep":         {Data: []byte{}},
		"specs/themes/descriptors.yml":                    {Data: []byte(themesYAML)},
	}
}

func loadFixtureDataset(t *testing.T) descriptors.Dataset {
	t.Helper()
	dataset, err := descriptors.Load(context.Background(), descriptorFixtureFS(), descriptors.LoadOptions{})
	if err != nil {
		t.Fatalf("descriptors.Load(fixture) error = %v", err)
	}
	return dataset
}

func TestRegistryExposesDescriptorCorpora(t *testing.T) {
	got := descriptors.Registry()
	want := []descriptors.Corpus{
		{Name: "cefr", Roots: []string{"specs/cefr"}, PathFields: []descriptors.Field{descriptors.FieldDomain, descriptors.FieldSubdomain}, DefaultCoverageAxes: []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldDomain, descriptors.FieldSubdomain, descriptors.FieldLevel}},
		{Name: "french", Roots: []string{"specs/french"}, PathFields: []descriptors.Field{descriptors.FieldDomain}, DefaultCoverageAxes: []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldDomain, descriptors.FieldLevel}},
		{Name: "texts", Roots: []string{"specs/texts"}, PathFields: []descriptors.Field{descriptors.FieldDomain}, DefaultCoverageAxes: []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldDomain, descriptors.FieldLevel}},
		{Name: "themes", Files: []string{"specs/themes/descriptors.yml"}, PathFields: []descriptors.Field{}, DefaultCoverageAxes: []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldLevel}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("descriptors.Registry() mismatch (-want +got):\n%s", diff)
	}
}

func TestCanonicalLevelsUseCEFROrder(t *testing.T) {
	got := descriptors.CanonicalLevels()
	want := []string{"pre_a1", "a1", "a2", "b1", "b2", "c1", "c2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("descriptors.CanonicalLevels() mismatch (-want +got):\n%s", diff)
	}
}

func TestNormalizeFiltersDeduplicatesAndPreservesIDCase(t *testing.T) {
	got := descriptors.NormalizeFilters(descriptors.Filters{
		Corpora:   []string{" CEFR ", "cefr", "Themes"},
		Domain:    " Production ",
		Subdomain: " Speaking ",
		Scales:    []string{" Addressing_Audiences ", "addressing_audiences"},
		Levels:    []string{" pre_a1 ", "pre-a1", "a1"},
		Code:      " deliver_toast ",
		ID:        "  Cefr.ID.Keeps.Case  ",
	})
	want := descriptors.Filters{
		Corpora:   []string{"cefr", "themes"},
		Domain:    "production",
		Subdomain: "speaking",
		Scales:    []string{"addressing_audiences"},
		Levels:    []string{"pre_a1", "a1"},
		Code:      "deliver_toast",
		ID:        "Cefr.ID.Keeps.Case",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("descriptors.NormalizeFilters() mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadMapsPathsNormalizesRowsSortsCanonicallyAndUsesYMLOnly(t *testing.T) {
	dataset := loadFixtureDataset(t)
	want := descriptors.Dataset{
		Scales: []descriptors.DescriptorScaleRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "addressing_audiences", ID: "cefr.production.speaking.descriptors.addressing_audiences", Description: []string{"Address audiences."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "turntaking", ID: "cefr.production.speaking.descriptors.turntaking", Description: []string{"Manage turntaking."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Code: "personal_identity", ID: "themes.descriptors.personal_identity", Description: []string{"Personal identity themes."}, File: "specs/themes/descriptors.yml"},
		},
		Descriptors: []descriptors.DescriptorRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a1", Code: "deliver_toast", ID: "cefr.production.speaking.descriptors.addressing_audiences.deliver_toast.a1", Description: "Can deliver a short toast.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a2", Code: "present_simple_announcement", ID: "cefr.production.speaking.descriptors.addressing_audiences.present_simple_announcement.a2", Description: "Can present a simple announcement.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "b1", Code: "maintain_exchange", ID: "cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1", Description: "Can maintain an exchange.\nCan add relevant details.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Scale: "personal_identity", Level: "pre_a1", Code: "personal_details", ID: "themes.descriptors.personal_identity.personal_details.pre_a1", Description: "Can share basic personal details.", File: "specs/themes/descriptors.yml"},
		},
	}
	if diff := cmp.Diff(want, dataset); diff != "" {
		t.Errorf("descriptors.Load(fixture) mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadReportsStructuredErrorsForMalformedYAMLAndInvalidRows(t *testing.T) {
	tests := []struct {
		name     string
		files    fstest.MapFS
		wantCode string
	}{
		{name: "malformed yaml", files: fstest.MapFS{"specs/cefr/production/speaking/descriptors.yml": {Data: []byte("id: broken\ncatalog: [")}}, wantCode: "invalid_yaml"},
		{name: "invalid row", files: fstest.MapFS{"specs/cefr/production/speaking/descriptors.yml": {Data: []byte("catalog:\n  - id: missing-code\nentries:\n  - scale: x\n")}}, wantCode: "invalid_row"},
		{name: "missing specs", files: fstest.MapFS{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := descriptors.Load(context.Background(), tt.files, descriptors.LoadOptions{Corpora: []string{"cefr"}})
			if err == nil {
				t.Fatal("descriptors.Load() error = nil, want structured error")
			}
			var coded descriptors.CodedError
			if !errors.As(err, &coded) {
				t.Fatalf("descriptors.Load() error = %T %[1]v, want descriptors.CodedError", err)
			}
			wantCode := tt.wantCode
			if wantCode == "" {
				wantCode = "missing_specs"
			}
			if coded.Code != wantCode {
				t.Errorf("coded error code = %q, want %q", coded.Code, wantCode)
			}
		})
	}
}

func TestEmbeddedSpecsContractCounts(t *testing.T) {
	dataset, err := descriptors.Load(context.Background(), specs.Bundle(), descriptors.LoadOptions{})
	if err != nil {
		t.Fatalf("descriptors.Load(embedded specs) error = %v", err)
	}

	counts := map[string]int{}
	for _, descriptor := range dataset.Descriptors {
		counts[descriptor.Corpus]++
	}
	wantCounts := map[string]int{"cefr": 1086, "french": 696, "texts": 355, "themes": 560}
	if diff := cmp.Diff(wantCounts, counts); diff != "" {
		t.Errorf("embedded descriptor counts mismatch (-want +got):\n%s", diff)
	}
	if got, want := len(dataset.Scales), 244; got != want {
		t.Errorf("embedded scale count = %d, want %d", got, want)
	}
	if got, want := len(dataset.Descriptors), 2697; got != want {
		t.Errorf("embedded descriptor count = %d, want %d", got, want)
	}
}
