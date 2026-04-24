package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestQueryScalesRequiresCorpusAndReturnsScaleRecords(t *testing.T) {
	got, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Domain: "production", Query: "audiences"})
	if err != nil {
		t.Fatalf("descriptors.QueryScales() error = %v", err)
	}
	want := descriptors.ScaleQueryResult{
		Kind:          "descriptor_scales",
		SchemaVersion: descriptors.SchemaVersion,
		Total:         1,
		Returned:      1,
		Offset:        0,
		Items: []descriptors.DescriptorScaleRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "addressing_audiences", ID: "cefr.production.speaking.descriptors.addressing_audiences", Description: []string{"Address audiences."}, File: "specs/cefr/production/speaking/descriptors.yml"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("descriptors.QueryScales() mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryScalesAppliesPathScaleAndIDFilters(t *testing.T) {
	got, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{
		Corpora:   []string{"cefr"},
		Domain:    "production",
		Subdomain: "speaking",
		Scales:    []string{"turntaking"},
		ID:        "cefr.production.speaking.descriptors.turntaking",
	})
	if err != nil {
		t.Fatalf("descriptors.QueryScales(filtered) error = %v", err)
	}
	wantCodes := []string{"turntaking"}
	codes := make([]string, 0, len(got.Items))
	for _, item := range got.Items {
		codes = append(codes, item.Code)
	}
	if diff := cmp.Diff(wantCodes, codes); diff != "" {
		t.Errorf("descriptors.QueryScales(filtered) codes mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryScalesSortsByRequestedFieldAndOrder(t *testing.T) {
	tests := []struct {
		name     string
		sortBy   descriptors.Field
		order    string
		wantCode []string
	}{
		{name: "code ascending", sortBy: descriptors.FieldCode, order: "asc", wantCode: []string{"addressing_audiences", "turntaking"}},
		{name: "code descending", sortBy: descriptors.FieldCode, order: "desc", wantCode: []string{"turntaking", "addressing_audiences"}},
		{name: "id descending", sortBy: descriptors.FieldID, order: "desc", wantCode: []string{"turntaking", "addressing_audiences"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, SortBy: tt.sortBy, SortOrder: tt.order})
			if err != nil {
				t.Fatalf("descriptors.QueryScales() error = %v", err)
			}
			codes := make([]string, 0, len(got.Items))
			for _, item := range got.Items {
				codes = append(codes, item.Code)
			}
			if diff := cmp.Diff(tt.wantCode, codes); diff != "" {
				t.Errorf("descriptors.QueryScales() codes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestQueryScalesPaginationNormalizesOffsets(t *testing.T) {
	got, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Offset: -1, Limit: 1})
	if err != nil {
		t.Fatalf("descriptors.QueryScales(negative offset) error = %v", err)
	}
	if got.Offset != 0 || got.Returned != 1 {
		t.Errorf("descriptors.QueryScales(negative offset) offset/returned = %d/%d, want 0/1", got.Offset, got.Returned)
	}

	got, err = descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Offset: 99, Limit: 1})
	if err != nil {
		t.Fatalf("descriptors.QueryScales(offset beyond length) error = %v", err)
	}
	if got.Offset != 99 || got.Returned != 0 || len(got.Items) != 0 {
		t.Errorf("descriptors.QueryScales(offset beyond length) = offset %d returned %d items %d, want 99/0/0", got.Offset, got.Returned, len(got.Items))
	}
}
