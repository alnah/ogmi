package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestSchemaReportsUnavailableFieldWithAvailableFields(t *testing.T) {
	_, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{
		Field:   descriptors.FieldSubdomain,
		Filters: descriptors.Filters{Corpora: []string{"themes"}},
	})
	coded := requireCodedError(t, err, "unavailable_field")

	wantInvalidFilter := descriptors.InvalidFilter{Field: "field", Value: "subdomain"}
	if diff := cmp.Diff(wantInvalidFilter, coded.Details.InvalidFilter); diff != "" {
		t.Errorf("unavailable field invalidFilter mismatch (-want +got):\n%s", diff)
	}
	if fieldsContain(coded.Details.AvailableFields, descriptors.FieldSubdomain) {
		t.Errorf("availableFields = %v, want subdomain omitted for themes", coded.Details.AvailableFields)
	}
	for _, wantField := range themeAvailableFields() {
		if !fieldsContain(coded.Details.AvailableFields, wantField) {
			t.Errorf("availableFields = %v, want it to include %q", coded.Details.AvailableFields, wantField)
		}
	}
}

func TestSchemaSummaryValuesAndFieldErrors(t *testing.T) {
	summary, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{})
	if err != nil {
		t.Fatalf("descriptors.Schema(summary) error = %v", err)
	}
	if summary.Kind != "descriptor_schema" || len(summary.Fields) == 0 {
		t.Errorf("descriptors.Schema(summary) = %+v, want descriptor_schema with fields", summary)
	}

	values, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{
		Field: descriptors.FieldLevel,
		Filters: descriptors.Filters{
			Corpora: []string{"cefr"},
		},
	})
	if err != nil {
		t.Fatalf("descriptors.Schema(values) error = %v", err)
	}
	wantValues := []string{"a1", "a2", "b1"}
	if values.Kind != "descriptor_schema_values" || !cmp.Equal(wantValues, values.Values) {
		t.Errorf("descriptors.Schema(values) = %+v, want kind descriptor_schema_values values %v", values, wantValues)
	}

	_, err = descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{
		Field: descriptors.Field("subdomian"),
	})
	coded := requireCodedError(t, err, "unknown_field")
	if diff := cmp.Diff([]string{"subdomain"}, coded.Suggestions); diff != "" {
		t.Errorf("unknown field suggestions mismatch (-want +got):\n%s", diff)
	}
	wantDetails := descriptors.ErrorDetails{
		InvalidFilter:   descriptors.InvalidFilter{Field: "field", Value: "subdomian"},
		AvailableFields: allDescriptorFields(),
	}
	if diff := cmp.Diff(wantDetails, coded.Details); diff != "" {
		t.Errorf("unknown field details mismatch (-want +got):\n%s", diff)
	}
}

func TestSchemaFieldValuesRespectFilters(t *testing.T) {
	values, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{
		Field: descriptors.FieldCode,
		Filters: descriptors.Filters{
			Corpora: []string{"cefr"},
			Domain:  "production",
			Scales:  []string{"turntaking"},
		},
	})
	if err != nil {
		t.Fatalf("descriptors.Schema(filtered values) error = %v", err)
	}
	wantValues := []string{"maintain_exchange", "open_simple_exchange"}
	if diff := cmp.Diff(wantValues, values.Values); diff != "" {
		t.Errorf("descriptors.Schema(filtered values) mismatch (-want +got):\n%s", diff)
	}
}

func themeAvailableFields() []descriptors.Field {
	return []descriptors.Field{
		descriptors.FieldCorpus,
		descriptors.FieldScale,
		descriptors.FieldLevel,
		descriptors.FieldCode,
		descriptors.FieldID,
	}
}

func allDescriptorFields() []descriptors.Field {
	return []descriptors.Field{
		descriptors.FieldCorpus,
		descriptors.FieldDomain,
		descriptors.FieldSubdomain,
		descriptors.FieldScale,
		descriptors.FieldLevel,
		descriptors.FieldCode,
		descriptors.FieldID,
	}
}
