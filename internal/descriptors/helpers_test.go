package descriptors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alnah/ogmi/internal/descriptors"
)

const (
	cefrSpeakingDescriptorsFile = "specs/cefr/production/speaking/descriptors.yml"
	themesDescriptorsFile       = "specs/themes/descriptors.yml"

	cefrProductionSpeakingDescriptorPrefix = "cefr.production.speaking.descriptors."
	addressingAudiencesScaleID             = cefrProductionSpeakingDescriptorPrefix + "addressing_audiences"
	turntakingScaleID                      = cefrProductionSpeakingDescriptorPrefix + "turntaking"
	deliverToastDescriptorID               = addressingAudiencesScaleID + ".deliver_toast.a1"
	presentSimpleAnnouncementDescriptorID  = addressingAudiencesScaleID + ".present_simple_announcement.a2"
	openSimpleExchangeDescriptorID         = turntakingScaleID + ".open_simple_exchange.a1"
	maintainExchangeDescriptorID           = turntakingScaleID + ".maintain_exchange.b1"

	themesDescriptorPrefix            = "themes.descriptors."
	personalIdentityScaleID           = themesDescriptorPrefix + "personal_identity"
	personalDetailsDescriptorID       = personalIdentityScaleID + ".personal_details.pre_a1"
	shareBasicIdentityFactsDescriptor = personalIdentityScaleID + ".share_basic_identity_facts.a1"
)

func queryDataset() descriptors.Dataset {
	return descriptors.Dataset{
		Scales: []descriptors.DescriptorScaleRecord{
			{
				Corpus:      "cefr",
				Domain:      "production",
				Subdomain:   "speaking",
				Code:        "addressing_audiences",
				ID:          addressingAudiencesScaleID,
				Description: []string{"Address audiences."},
				File:        cefrSpeakingDescriptorsFile,
			},
			{
				Corpus:      "cefr",
				Domain:      "production",
				Subdomain:   "speaking",
				Code:        "turntaking",
				ID:          turntakingScaleID,
				Description: []string{"Manage turntaking."},
				File:        cefrSpeakingDescriptorsFile,
			},
			{
				Corpus:      "themes",
				Code:        "personal_identity",
				ID:          personalIdentityScaleID,
				Description: []string{"Personal identity themes."},
				File:        themesDescriptorsFile,
			},
		},
		Descriptors: []descriptors.DescriptorRecord{
			{
				Corpus:      "cefr",
				Domain:      "production",
				Subdomain:   "speaking",
				Scale:       "addressing_audiences",
				Level:       "a1",
				Code:        "deliver_toast",
				ID:          deliverToastDescriptorID,
				Description: "Can deliver a short toast.",
				File:        cefrSpeakingDescriptorsFile,
			},
			{
				Corpus:      "cefr",
				Domain:      "production",
				Subdomain:   "speaking",
				Scale:       "addressing_audiences",
				Level:       "a2",
				Code:        "present_simple_announcement",
				ID:          presentSimpleAnnouncementDescriptorID,
				Description: "Can present a simple announcement.",
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
				Corpus:      "themes",
				Scale:       "personal_identity",
				Level:       "a1",
				Code:        "share_basic_identity_facts",
				ID:          shareBasicIdentityFactsDescriptor,
				Description: "Can share basic identity facts.",
				File:        themesDescriptorsFile,
			},
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
