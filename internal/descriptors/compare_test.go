package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestCompareLevelsReturnsCompleteSummariesBeforeLimit(t *testing.T) {
	got, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr", Scale: "turntaking", Levels: []string{"a1", "a2", "b1"}, Limit: 1})
	if err != nil {
		t.Fatalf("descriptors.CompareLevels() error = %v", err)
	}

	if got.Kind != "descriptor_level_comparison" || got.SchemaVersion != descriptors.SchemaVersion || got.Total != 2 || got.Returned != 1 {
		t.Errorf("descriptors.CompareLevels() metadata = kind %q schema %q total %d returned %d, want descriptor_level_comparison %s 2 1", got.Kind, got.SchemaVersion, got.Total, got.Returned, descriptors.SchemaVersion)
	}
	wantLevels := []string{"a1", "a2", "b1"}
	if diff := cmp.Diff(wantLevels, got.Levels); diff != "" {
		t.Errorf("descriptors.CompareLevels() levels mismatch (-want +got):\n%s", diff)
	}
	wantSummaries := []descriptors.LevelSummary{{Level: "a1", Total: 1, Present: true}, {Level: "a2", Total: 0, Present: false}, {Level: "b1", Total: 1, Present: true}}
	if diff := cmp.Diff(wantSummaries, got.Summaries); diff != "" {
		t.Errorf("descriptors.CompareLevels() summaries mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"a2"}, got.MissingLevels); diff != "" {
		t.Errorf("descriptors.CompareLevels() missing levels mismatch (-want +got):\n%s", diff)
	}
	if gotItems := len(got.Items); gotItems != 1 {
		t.Errorf("descriptors.CompareLevels() items len = %d, want limit to narrow returned items only", gotItems)
	}
}

func TestCompareLevelsDerivesPresentLevelsWhenOmitted(t *testing.T) {
	got, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr", Scale: "turntaking"})
	if err != nil {
		t.Fatalf("descriptors.CompareLevels(levels omitted) error = %v", err)
	}

	wantLevels := []string{"a1", "b1"}
	if diff := cmp.Diff(wantLevels, got.Levels); diff != "" {
		t.Errorf("descriptors.CompareLevels(levels omitted) levels mismatch (-want +got):\n%s", diff)
	}
	wantSummaries := []descriptors.LevelSummary{{Level: "a1", Total: 1, Present: true}, {Level: "b1", Total: 1, Present: true}}
	if diff := cmp.Diff(wantSummaries, got.Summaries); diff != "" {
		t.Errorf("descriptors.CompareLevels(levels omitted) summaries mismatch (-want +got):\n%s", diff)
	}
	if len(got.MissingLevels) != 0 {
		t.Errorf("descriptors.CompareLevels(levels omitted) missing = %v, want none", got.MissingLevels)
	}
}
