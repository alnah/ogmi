package descriptors

import (
	"context"
	"sort"
)

// CompareLevels summarizes descriptor coverage across levels for one scale.
func CompareLevels(ctx context.Context, dataset Dataset, input CompareLevelsInput) (CompareLevelsResult, error) {
	if err := ctx.Err(); err != nil {
		return CompareLevelsResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	corpus := normalizeToken(input.Corpus)
	scale := normalizeToken(input.Scale)
	if corpus == "" || scale == "" {
		return CompareLevelsResult{}, CodedError{Code: "missing_required_filter", Message: "compare-levels requires corpus and scale"}
	}
	if !knownCorpus(corpus) {
		return CompareLevelsResult{}, unknownCorpusError(corpus)
	}
	filters := NormalizeFilters(Filters{Corpora: []string{corpus}, Domain: input.Domain, Subdomain: input.Subdomain, Scales: []string{scale}, Levels: input.Levels, Query: input.Query})
	items := filterDescriptors(dataset.Descriptors, filters)
	sort.SliceStable(items, func(i, j int) bool { return compareDescriptor(items[i], items[j]) < 0 })
	levels := filters.Levels
	if len(levels) == 0 {
		levels = levelsFromItems(items)
	}
	summaries := make([]LevelSummary, 0, len(levels))
	missing := make([]string, 0)
	for _, level := range levels {
		count := 0
		for _, item := range items {
			if item.Level == level {
				count++
			}
		}
		present := count > 0
		if !present {
			missing = append(missing, level)
		}
		summaries = append(summaries, LevelSummary{Level: level, Total: count, Present: present})
	}
	paged := items
	if input.Limit > 0 && input.Limit < len(paged) {
		paged = paged[:input.Limit]
	}
	return CompareLevelsResult{Kind: "descriptor_level_comparison", SchemaVersion: SchemaVersion, Levels: levels, Summaries: summaries, MissingLevels: missing, Total: len(items), Returned: len(paged), Items: paged}, nil
}

func levelsFromItems(items []DescriptorRecord) []string {
	seen := make(map[string]bool)
	levels := []string{}
	for _, item := range items {
		if !seen[item.Level] {
			seen[item.Level] = true
			levels = append(levels, item.Level)
		}
	}
	sort.SliceStable(levels, func(i, j int) bool { return compareLevels(levels[i], levels[j]) < 0 })
	return levels
}
