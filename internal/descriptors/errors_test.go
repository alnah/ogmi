package descriptors_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestQueryReportsTypedInvalidFilterDetails(t *testing.T) {
	_, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Levels: []string{"Z9"}})
	coded := requireCodedError(t, err, "invalid_filter")
	wantInvalidFilter := descriptors.InvalidFilter{Field: "level", Value: "Z9"}
	if diff := cmp.Diff(wantInvalidFilter, coded.Details.InvalidFilter); diff != "" {
		t.Errorf("invalidFilter mismatch (-want +got):\n%s", diff)
	}
}

func TestUnknownCorpusSuggestionsUseCloseMatchesOnly(t *testing.T) {
	tests := []struct {
		name            string
		corpus          string
		wantSuggestions []string
	}{
		{name: "close typo", corpus: "cefrr", wantSuggestions: []string{"cefr"}},
		{name: "distant typo", corpus: "zzzzzz", wantSuggestions: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{tt.corpus}})
			coded := requireCodedError(t, err, "unknown_corpus")
			if diff := cmp.Diff(tt.wantSuggestions, coded.Suggestions); diff != "" {
				t.Errorf("unknown corpus suggestions mismatch (-want +got):\n%s", diff)
			}
			wantInvalidFilter := descriptors.InvalidFilter{Field: "corpus", Value: tt.corpus}
			if diff := cmp.Diff(wantInvalidFilter, coded.Details.InvalidFilter); diff != "" {
				t.Errorf("unknown corpus invalidFilter mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDescriptorOperationsReturnContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "query", run: func() error {
			_, err := descriptors.Query(ctx, queryDataset(), descriptors.Filters{})
			return err
		}},
		{name: "scales", run: func() error {
			_, err := descriptors.QueryScales(ctx, queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}})
			return err
		}},
		{name: "get", run: func() error {
			_, err := descriptors.Get(ctx, queryDataset(), descriptors.GetInput{Corpus: "cefr", Code: "deliver_toast"})
			return err
		}},
		{name: "compare", run: func() error {
			_, err := descriptors.CompareLevels(ctx, queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr", Scale: "turntaking"})
			return err
		}},
		{name: "coverage", run: func() error {
			_, err := descriptors.Coverage(ctx, queryDataset(), descriptors.CoverageInput{Corpus: "cefr"})
			return err
		}},
		{name: "schema", run: func() error {
			_, err := descriptors.Schema(ctx, queryDataset(), descriptors.SchemaInput{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = requireCodedError(t, tt.run(), "internal")
		})
	}
}

func TestDescriptorSadPathsReturnStableCodedErrors(t *testing.T) {
	tests := []struct {
		name     string
		run      func() error
		wantCode string
	}{
		{name: "scales without corpus", run: func() error {
			_, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "get without id or code", run: func() error {
			_, err := descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{Corpus: "cefr"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "compare without corpus", run: func() error {
			_, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Scale: "turntaking"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "compare without scale", run: func() error {
			_, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "coverage without corpus", run: func() error {
			_, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "unknown corpus", run: func() error {
			_, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefrr"}})
			return err
		}, wantCode: "unknown_corpus"},
		{name: "missing requested specs", run: func() error {
			_, err := descriptors.Load(context.Background(), fstest.MapFS{}, descriptors.LoadOptions{Corpora: []string{"themes"}})
			return err
		}, wantCode: "missing_specs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = requireCodedError(t, tt.run(), tt.wantCode)
		})
	}
}
