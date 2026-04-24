package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestGetPrefersIDAndReportsAmbiguousCode(t *testing.T) {
	got, err := descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{
		Corpus: "cefr",
		Code:   "deliver_toast",
		ID:     openSimpleExchangeDescriptorID,
	})
	if err != nil {
		t.Fatalf("descriptors.Get(id wins) error = %v", err)
	}
	if got.Kind != "descriptor_get" || got.Descriptor.Code != "open_simple_exchange" {
		t.Errorf("descriptors.Get(id wins) = %+v, want open_simple_exchange descriptor_get", got)
	}

	_, err = descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{
		Corpus: "cefr",
		Code:   "open_simple_exchange",
	})
	if err != nil {
		t.Fatalf("descriptors.Get(unique code) error = %v", err)
	}

	ambiguousDataset := queryDataset()
	ambiguousDataset.Descriptors = append(ambiguousDataset.Descriptors, descriptors.DescriptorRecord{
		Corpus:      "cefr",
		Domain:      "interaction",
		Subdomain:   "speaking",
		Scale:       "turntaking",
		Level:       "a2",
		Code:        "open_simple_exchange",
		ID:          "cefr.interaction.speaking.descriptors.turntaking.open_simple_exchange.a2",
		Description: "Can open another exchange.",
		File:        "specs/cefr/interaction/speaking/descriptors.yml",
	})
	_, err = descriptors.Get(context.Background(), ambiguousDataset, descriptors.GetInput{
		Corpus: "cefr",
		Code:   "open_simple_exchange",
	})
	_ = requireCodedError(t, err, "ambiguous_lookup")
}

func TestGetReturnsNotFoundAndPreservesIDCaseSemantics(t *testing.T) {
	id := deliverToastDescriptorID
	got, err := descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{Corpus: "cefr", ID: id})
	if err != nil {
		t.Fatalf("descriptors.Get(exact id) error = %v", err)
	}
	if got.Descriptor.ID != id {
		t.Errorf("descriptors.Get(exact id) id = %q, want %q", got.Descriptor.ID, id)
	}
	wantDescription := []string{"Can deliver a short toast."}
	if diff := cmp.Diff(wantDescription, got.Description); diff != "" {
		t.Errorf("descriptors.Get(exact id) description mismatch (-want +got):\n%s", diff)
	}

	_, err = descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{
		Corpus: "cefr",
		ID:     "CEFR.PRODUCTION.SPEAKING.DESCRIPTORS.ADDRESSING_AUDIENCES.DELIVER_TOAST.A1",
	})
	_ = requireCodedError(t, err, "not_found")

	_, err = descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{
		Corpus: "cefr",
		Code:   "does_not_exist",
	})
	_ = requireCodedError(t, err, "not_found")
}
