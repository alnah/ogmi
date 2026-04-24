package descriptors

import "strings"

// NormalizeFilters trims, lowercases, normalizes levels, and deduplicates filters.
func NormalizeFilters(filters Filters) Filters {
	filters.Corpora = normalizeTokens(filters.Corpora)
	filters.Domain = normalizeToken(filters.Domain)
	filters.Subdomain = normalizeToken(filters.Subdomain)
	filters.Scales = normalizeTokens(filters.Scales)
	filters.Levels = normalizeLevels(filters.Levels)
	filters.Code = normalizeToken(filters.Code)
	filters.ID = strings.TrimSpace(filters.ID)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.SortBy = Field(normalizeToken(string(filters.SortBy)))
	filters.SortOrder = normalizeToken(filters.SortOrder)
	filters.GroupBy = normalizeFields(filters.GroupBy)
	return filters
}

func normalizeTokens(values []string) []string {
	out := []string{}
	seen := make(map[string]bool)
	for _, value := range values {
		normalized := normalizeToken(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeLevels(values []string) []string {
	out := []string{}
	seen := make(map[string]bool)
	for _, value := range values {
		normalized := normalizeLevel(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeFields(values []Field) []Field {
	if len(values) == 0 {
		return nil
	}
	out := []Field{}
	seen := make(map[Field]bool)
	for _, value := range values {
		normalized := Field(normalizeToken(string(value)))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeToken(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func normalizeLevel(value string) string {
	normalized := normalizeToken(value)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	if strings.ReplaceAll(normalized, " ", "_") == "pre_a1" || strings.ReplaceAll(normalized, "_", " ") == "pre a1" {
		return "pre_a1"
	}
	return normalized
}

func normalizeDescription(lines []string) []string {
	out := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
