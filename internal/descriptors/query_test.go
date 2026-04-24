package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestQueryFiltersSortsPaginatesGroupsAndFacets(t *testing.T) {
	got, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{
		Corpora:    []string{"CEFR"},
		Domain:     "Production",
		Subdomain:  "Speaking",
		Scales:     []string{"turntaking", "addressing_audiences"},
		Levels:     []string{"b1", "a1"},
		Offset:     1,
		Limit:      2,
		SortBy:     descriptors.FieldCode,
		SortOrder:  "asc",
		GroupBy:    []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel},
		WithFacets: true,
	})
	if err != nil {
		t.Fatalf("descriptors.Query() error = %v", err)
	}

	wantItems := []descriptors.DescriptorRecord{
		{
			Corpus:      "cefr",
			Domain:      "production",
			Subdomain:   "speaking",
			Scale:       "turntaking",
			Level:       "b1",
			Code:        "maintain_exchange",
			ID:          maintainExchangeDescriptorID,
			Description: "Can maintain an exchange.",
			File:        cefrSpeakingDescriptorsFile,
		},
		{
			Corpus:      "cefr",
			Domain:      "production",
			Subdomain:   "speaking",
			Scale:       "turntaking",
			Level:       "a1",
			Code:        "open_simple_exchange",
			ID:          openSimpleExchangeDescriptorID,
			Description: "Can open a simple exchange.",
			File:        cefrSpeakingDescriptorsFile,
		},
	}
	if got.Kind != "descriptor_list" || got.SchemaVersion != descriptors.SchemaVersion ||
		got.Total != 3 || got.Returned != 2 || got.Offset != 1 {
		t.Errorf(
			"descriptors.Query() metadata = kind %q schema %q total %d returned %d offset %d, want %s %s 3 2 1",
			got.Kind,
			got.SchemaVersion,
			got.Total,
			got.Returned,
			got.Offset,
			"descriptor_list",
			descriptors.SchemaVersion,
		)
	}
	if diff := cmp.Diff(wantItems, got.Items); diff != "" {
		t.Errorf("descriptors.Query() items mismatch (-want +got):\n%s", diff)
	}

	wantScaleBuckets := []descriptors.FacetBucket{
		{Value: "addressing_audiences", Count: 1},
		{Value: "turntaking", Count: 2},
	}
	if diff := cmp.Diff(wantScaleBuckets, facetBuckets(got.Facets, descriptors.FieldScale)); diff != "" {
		t.Errorf("descriptors.Query() scale facet buckets mismatch (-want +got):\n%s", diff)
	}
	wantLevelBuckets := []descriptors.FacetBucket{{Value: "a1", Count: 2}, {Value: "b1", Count: 1}}
	if diff := cmp.Diff(wantLevelBuckets, facetBuckets(got.Facets, descriptors.FieldLevel)); diff != "" {
		t.Errorf("descriptors.Query() level facet buckets mismatch (-want +got):\n%s", diff)
	}

	wantGroups := []descriptorGroupSummary{
		{
			KeyFields: []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel},
			Scale:     "turntaking",
			Level:     "b1",
			Total:     1,
			ItemIDs:   []string{maintainExchangeDescriptorID},
		},
		{
			KeyFields: []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel},
			Scale:     "turntaking",
			Level:     "a1",
			Total:     1,
			ItemIDs:   []string{openSimpleExchangeDescriptorID},
		},
	}
	if diff := cmp.Diff(wantGroups, groupSummaries(got.Groups)); diff != "" {
		t.Errorf("descriptors.Query() groups mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryMatchesTextAcrossDescriptorFields(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantIDs   []string
		wantTotal int
	}{
		{
			name:      "description",
			query:     "short toast",
			wantIDs:   []string{deliverToastDescriptorID},
			wantTotal: 1,
		},
		{
			name:      "id",
			query:     "present_simple_announcement.a2",
			wantIDs:   []string{presentSimpleAnnouncementDescriptorID},
			wantTotal: 1,
		},
		{
			name:  "path",
			query: "production",
			wantIDs: []string{
				deliverToastDescriptorID,
				presentSimpleAnnouncementDescriptorID,
				openSimpleExchangeDescriptorID,
				maintainExchangeDescriptorID,
			},
			wantTotal: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requireQuery(t, descriptors.Filters{Corpora: []string{"cefr"}, Query: tt.query})
			if got.Total != tt.wantTotal {
				t.Fatalf("descriptors.Query(query %q) total = %d, want %d", tt.query, got.Total, tt.wantTotal)
			}
			if diff := cmp.Diff(tt.wantIDs, descriptorIDs(got.Items)); diff != "" {
				t.Errorf("descriptors.Query(query %q) IDs mismatch (-want +got):\n%s", tt.query, diff)
			}
		})
	}
}

func TestQueryPaginationNormalizesOffsets(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		wantResultOffs int
		wantIDs        []string
	}{
		{name: "negative", offset: -3, wantResultOffs: 0, wantIDs: []string{deliverToastDescriptorID}},
		{name: "beyond length", offset: 99, wantResultOffs: 99, wantIDs: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requireQuery(t, descriptors.Filters{
				Corpora: []string{"cefr"},
				SortBy:  descriptors.FieldID,
				Offset:  tt.offset,
				Limit:   1,
			})
			if got.Offset != tt.wantResultOffs {
				t.Errorf("descriptors.Query(offset %d) offset = %d, want %d", tt.offset, got.Offset, tt.wantResultOffs)
			}
			if diff := cmp.Diff(tt.wantIDs, descriptorIDs(got.Items)); diff != "" {
				t.Errorf("descriptors.Query(offset %d) IDs mismatch (-want +got):\n%s", tt.offset, diff)
			}
		})
	}
}
