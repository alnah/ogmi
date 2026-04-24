package descriptors

import "sort"

func buildFacets(items []DescriptorRecord) []Facet {
	fields := []Field{FieldCorpus, FieldDomain, FieldSubdomain, FieldScale, FieldLevel, FieldCode, FieldID}
	facets := make([]Facet, 0, len(fields))
	for _, field := range fields {
		counts := make(map[string]int)
		for _, item := range items {
			value := recordField(item, field)
			if value != "" {
				counts[value]++
			}
		}
		if len(counts) == 0 {
			continue
		}
		values := make([]string, 0, len(counts))
		for value := range counts {
			values = append(values, value)
		}
		sort.SliceStable(values, func(i, j int) bool {
			if field == FieldLevel {
				return compareLevels(values[i], values[j]) < 0
			}
			return values[i] < values[j]
		})
		buckets := make([]FacetBucket, 0, len(values))
		for _, value := range values {
			buckets = append(buckets, FacetBucket{Value: value, Count: counts[value]})
		}
		facets = append(facets, Facet{Field: field, Buckets: buckets})
	}
	return facets
}
