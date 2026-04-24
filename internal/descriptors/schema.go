package descriptors

import (
	"context"
	"fmt"
	"sort"
)

func Schema(ctx context.Context, dataset Dataset, input SchemaInput) (SchemaResult, error) {
	if err := ctx.Err(); err != nil {
		return SchemaResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	if input.Field == "" {
		return SchemaResult{Kind: "descriptor_schema", SchemaVersion: SchemaVersion, Fields: knownFields()}, nil
	}
	field := Field(normalizeToken(string(input.Field)))
	if !isKnownField(field) {
		return SchemaResult{}, CodedError{Code: "unknown_field", Message: fmt.Sprintf("Unknown descriptor field: %s", input.Field), Suggestions: suggestions(string(input.Field), fieldNames()), Details: ErrorDetails{InvalidFilter: InvalidFilter{Field: "field", Value: string(input.Field)}, AvailableFields: knownFields()}}
	}
	filters := NormalizeFilters(input.Filters)
	available := availableFieldsForFilters(filters)
	if !fieldIn(field, available) {
		return SchemaResult{}, CodedError{Code: "unavailable_field", Message: fmt.Sprintf("Descriptor field %s is unavailable", field), Details: ErrorDetails{InvalidFilter: InvalidFilter{Field: "field", Value: string(field)}, AvailableFields: available}}
	}
	items := filterDescriptors(dataset.Descriptors, filters)
	values := distinctFieldValues(items, field)
	return SchemaResult{Kind: "descriptor_schema_values", SchemaVersion: SchemaVersion, Values: values}, nil
}

func distinctFieldValues(items []DescriptorRecord, field Field) []string {
	seen := make(map[string]bool)
	values := []string{}
	for _, item := range items {
		value := recordField(item, field)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	sort.SliceStable(values, func(i, j int) bool {
		if field == FieldLevel {
			return compareLevels(values[i], values[j]) < 0
		}
		return values[i] < values[j]
	})
	return values
}

func knownFields() []Field {
	return []Field{FieldCorpus, FieldDomain, FieldSubdomain, FieldScale, FieldLevel, FieldCode, FieldID}
}

func fieldNames() []string {
	fields := knownFields()
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, string(field))
	}
	return names
}

func isKnownField(field Field) bool { return fieldIn(field, knownFields()) }

func availableFieldsForFilters(filters Filters) []Field {
	corpora := filters.Corpora
	if len(corpora) == 0 {
		return knownFields()
	}
	available := knownFields()
	for _, corpusName := range corpora {
		corpus, ok := corpusByName(corpusName)
		if !ok {
			continue
		}
		available = intersectFields(available, availableFieldsForCorpus(corpus))
	}
	return available
}

func availableFieldsForCorpus(corpus Corpus) []Field {
	fields := []Field{FieldCorpus}
	fields = append(fields, corpus.PathFields...)
	fields = append(fields, FieldScale, FieldLevel, FieldCode, FieldID)
	return fields
}

func intersectFields(left, right []Field) []Field {
	out := []Field{}
	for _, field := range left {
		if fieldIn(field, right) {
			out = append(out, field)
		}
	}
	return out
}
