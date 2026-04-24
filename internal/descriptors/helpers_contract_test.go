package descriptors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alnah/ogmi/internal/descriptors"
)

func queryDataset() descriptors.Dataset {
	return descriptors.Dataset{
		Scales: []descriptors.DescriptorScaleRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "addressing_audiences", ID: "cefr.production.speaking.descriptors.addressing_audiences", Description: []string{"Address audiences."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "turntaking", ID: "cefr.production.speaking.descriptors.turntaking", Description: []string{"Manage turntaking."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Code: "personal_identity", ID: "themes.descriptors.personal_identity", Description: []string{"Personal identity themes."}, File: "specs/themes/descriptors.yml"},
		},
		Descriptors: []descriptors.DescriptorRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a1", Code: "deliver_toast", ID: "cefr.production.speaking.descriptors.addressing_audiences.deliver_toast.a1", Description: "Can deliver a short toast.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a2", Code: "present_simple_announcement", ID: "cefr.production.speaking.descriptors.addressing_audiences.present_simple_announcement.a2", Description: "Can present a simple announcement.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "a1", Code: "open_simple_exchange", ID: "cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1", Description: "Can open a simple exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "b1", Code: "maintain_exchange", ID: "cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1", Description: "Can maintain an exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Scale: "personal_identity", Level: "a1", Code: "share_basic_identity_facts", ID: "themes.descriptors.personal_identity.share_basic_identity_facts.a1", Description: "Can share basic identity facts.", File: "specs/themes/descriptors.yml"},
		},
	}
}

type descriptorGroupSummary struct {
	KeyFields []descriptors.Field
	Scale     string
	Level     string
	Total     int
	ItemIDs   []string
}

func descriptorIDs(records []descriptors.DescriptorRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	return ids
}

func groupSummaries(groups []descriptors.Group) []descriptorGroupSummary {
	summaries := make([]descriptorGroupSummary, 0, len(groups))
	for _, group := range groups {
		summaries = append(summaries, descriptorGroupSummary{
			KeyFields: group.KeyFields,
			Scale:     group.Key[descriptors.FieldScale],
			Level:     group.Key[descriptors.FieldLevel],
			Total:     group.Total,
			ItemIDs:   descriptorIDs(group.Items),
		})
	}
	return summaries
}

func facetBuckets(facets []descriptors.Facet, field descriptors.Field) []descriptors.FacetBucket {
	for _, facet := range facets {
		if facet.Field == field {
			return facet.Buckets
		}
	}
	return nil
}

func coverageRow(rows []descriptors.CoverageRow, scale string) (descriptors.CoverageRow, bool) {
	for _, row := range rows {
		if row.Scale == scale {
			return row, true
		}
	}
	return descriptors.CoverageRow{}, false
}

func coverageCell(cells []descriptors.CoverageCell, level string) (descriptors.CoverageCell, bool) {
	for _, cell := range cells {
		if cell.Level == level {
			return cell, true
		}
	}
	return descriptors.CoverageCell{}, false
}

func fieldsContain(fields []descriptors.Field, want descriptors.Field) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}

func requireCodedError(t *testing.T, err error, wantCode string) descriptors.CodedError {
	t.Helper()
	var coded descriptors.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("error = %T %[1]v, want descriptors.CodedError", err)
	}
	if coded.Code != wantCode {
		t.Fatalf("coded error code = %q, want %q; error = %+v", coded.Code, wantCode, coded)
	}
	return coded
}

func requireQuery(t *testing.T, filters descriptors.Filters) descriptors.QueryResult {
	t.Helper()
	got, err := descriptors.Query(context.Background(), queryDataset(), filters)
	if err != nil {
		t.Fatalf("descriptors.Query() error = %v", err)
	}
	return got
}
