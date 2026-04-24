package descriptors

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Query returns descriptors matching filters, sorted and paginated.
func Query(ctx context.Context, dataset Dataset, filters Filters) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return QueryResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	if err := validateFilters(filters); err != nil {
		return QueryResult{}, err
	}
	filters = NormalizeFilters(filters)
	items := filterDescriptors(dataset.Descriptors, filters)
	sort.SliceStable(items, func(i, j int) bool {
		return compareDescriptorForQuery(items[i], items[j], filters.SortBy, filters.SortOrder) < 0
	})
	facets := []Facet(nil)
	if filters.WithFacets {
		facets = buildFacets(items)
	}
	paged := paginateDescriptors(items, filters.Offset, filters.Limit)
	groups := buildGroups(paged, filters.GroupBy)
	return QueryResult{Kind: "descriptor_list", SchemaVersion: SchemaVersion, Filters: filters, Total: len(items), Returned: len(paged), Offset: normalizedOffset(filters.Offset), Items: paged, Groups: groups, Facets: facets}, nil
}

// QueryScales returns scale records matching filters, sorted and paginated.
func QueryScales(ctx context.Context, dataset Dataset, filters Filters) (ScaleQueryResult, error) {
	if err := ctx.Err(); err != nil {
		return ScaleQueryResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	if len(normalizeTokens(filters.Corpora)) == 0 {
		return ScaleQueryResult{}, CodedError{Code: "missing_required_filter", Message: "descriptor scales requires corpus"}
	}
	if err := validateFilters(filters); err != nil {
		return ScaleQueryResult{}, err
	}
	filters = NormalizeFilters(filters)
	items := filterScales(dataset.Scales, filters)
	sort.SliceStable(items, func(i, j int) bool {
		return compareScaleForQuery(items[i], items[j], filters.SortBy, filters.SortOrder) < 0
	})
	paged := paginateScales(items, filters.Offset, filters.Limit)
	return ScaleQueryResult{Kind: "descriptor_scales", SchemaVersion: SchemaVersion, Total: len(items), Returned: len(paged), Offset: normalizedOffset(filters.Offset), Items: paged}, nil
}

// Get returns one descriptor selected by id or code.
func Get(ctx context.Context, dataset Dataset, input GetInput) (GetResult, error) {
	if err := ctx.Err(); err != nil {
		return GetResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	corpus := normalizeToken(input.Corpus)
	if corpus == "" {
		return GetResult{}, CodedError{Code: "missing_required_filter", Message: "get descriptor requires corpus"}
	}
	if !knownCorpus(corpus) {
		return GetResult{}, unknownCorpusError(corpus)
	}
	id := strings.TrimSpace(input.ID)
	code := normalizeToken(input.Code)
	if id == "" && code == "" {
		return GetResult{}, CodedError{Code: "missing_required_filter", Message: "get descriptor requires --id or --code"}
	}
	filters := Filters{Corpora: []string{corpus}, Domain: input.Domain, Subdomain: input.Subdomain, Scales: []string{input.Scale}, Levels: []string{input.Level}}
	filters = NormalizeFilters(filters)
	matches := make([]DescriptorRecord, 0)
	for _, descriptor := range dataset.Descriptors {
		if descriptor.Corpus != corpus || !matchesPathAndScale(descriptor, filters) {
			continue
		}
		if id != "" {
			if descriptor.ID == id {
				matches = append(matches, descriptor)
			}
			continue
		}
		if descriptor.Code == code {
			matches = append(matches, descriptor)
		}
	}
	if len(matches) == 0 {
		return GetResult{}, CodedError{Code: "not_found", Message: "Descriptor not found"}
	}
	if len(matches) > 1 {
		return GetResult{}, CodedError{Code: "ambiguous_lookup", Message: "Descriptor lookup is ambiguous. Provide --scale and --level or use --id."}
	}
	match := matches[0]
	return GetResult{Kind: "descriptor_get", SchemaVersion: SchemaVersion, Descriptor: match, Description: splitDescription(match.Description)}, nil
}

func validateFilters(filters Filters) error {
	for _, corpus := range normalizeTokens(filters.Corpora) {
		if !knownCorpus(corpus) {
			return unknownCorpusError(corpus)
		}
	}
	for _, level := range filters.Levels {
		normalized := normalizeLevel(level)
		if normalized != "" && !isCanonicalLevel(normalized) {
			return CodedError{Code: "invalid_filter", Message: fmt.Sprintf("Invalid descriptor level: %s", level), Details: ErrorDetails{InvalidFilter: InvalidFilter{Field: "level", Value: level}}}
		}
	}
	return nil
}

func filterDescriptors(records []DescriptorRecord, filters Filters) []DescriptorRecord {
	out := make([]DescriptorRecord, 0, len(records))
	for _, record := range records {
		if matchesDescriptor(record, filters) {
			out = append(out, record)
		}
	}
	return out
}

func matchesDescriptor(record DescriptorRecord, filters Filters) bool {
	if len(filters.Corpora) > 0 && !stringIn(record.Corpus, filters.Corpora) {
		return false
	}
	if !matchesPathAndScale(record, filters) {
		return false
	}
	if filters.Code != "" && record.Code != filters.Code {
		return false
	}
	if filters.ID != "" && record.ID != filters.ID {
		return false
	}
	if filters.Query != "" && !matchesText(record, filters.Query) {
		return false
	}
	return true
}

func matchesPathAndScale(record DescriptorRecord, filters Filters) bool {
	if filters.Domain != "" && record.Domain != filters.Domain {
		return false
	}
	if filters.Subdomain != "" && record.Subdomain != filters.Subdomain {
		return false
	}
	if len(filters.Scales) > 0 && !stringIn(record.Scale, filters.Scales) {
		return false
	}
	if len(filters.Levels) > 0 && !stringIn(record.Level, filters.Levels) {
		return false
	}
	return true
}

func filterScales(records []DescriptorScaleRecord, filters Filters) []DescriptorScaleRecord {
	out := make([]DescriptorScaleRecord, 0, len(records))
	for _, record := range records {
		if len(filters.Corpora) > 0 && !stringIn(record.Corpus, filters.Corpora) {
			continue
		}
		if filters.Domain != "" && record.Domain != filters.Domain {
			continue
		}
		if filters.Subdomain != "" && record.Subdomain != filters.Subdomain {
			continue
		}
		if len(filters.Scales) > 0 && !stringIn(record.Code, filters.Scales) {
			continue
		}
		if filters.ID != "" && record.ID != filters.ID {
			continue
		}
		if filters.Query != "" && !matchesScaleText(record, filters.Query) {
			continue
		}
		out = append(out, record)
	}
	return out
}

func matchesText(record DescriptorRecord, query string) bool {
	needle := strings.ToLower(query)
	values := []string{record.Corpus, record.Domain, record.Subdomain, record.Scale, record.Level, record.Code, record.ID, record.File, record.Description}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func matchesScaleText(record DescriptorScaleRecord, query string) bool {
	needle := strings.ToLower(query)
	values := []string{record.Corpus, record.Domain, record.Subdomain, record.Code, record.ID, record.File, strings.Join(record.Description, "\n")}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func paginateDescriptors(items []DescriptorRecord, offset, limit int) []DescriptorRecord {
	offset = normalizedOffset(offset)
	if offset >= len(items) {
		return []DescriptorRecord{}
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return append([]DescriptorRecord(nil), items[offset:end]...)
}

func paginateScales(items []DescriptorScaleRecord, offset, limit int) []DescriptorScaleRecord {
	offset = normalizedOffset(offset)
	if offset >= len(items) {
		return []DescriptorScaleRecord{}
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return append([]DescriptorScaleRecord(nil), items[offset:end]...)
}

func normalizedOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func splitDescription(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
