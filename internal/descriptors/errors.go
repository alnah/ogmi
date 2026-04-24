package descriptors

import "fmt"

// InvalidFilter identifies a rejected descriptor filter value.
type InvalidFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// ErrorDetails carries optional structured fields for coded errors.
type ErrorDetails struct {
	InvalidFilter   InvalidFilter `json:"invalidFilter,omitempty"`
	AvailableFields []Field       `json:"availableFields,omitempty"`
}

// CodedError describes a stable error shape for JSON output.
type CodedError struct {
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	Suggestions []string     `json:"suggestions,omitempty"`
	Details     ErrorDetails `json:"details,omitempty"`
}

func (e CodedError) Error() string { return e.Message }

func unknownCorpusError(value string) CodedError {
	return CodedError{Code: "unknown_corpus", Message: fmt.Sprintf("Unknown descriptor corpus: %s", value), Suggestions: suggestions(value, corpusNames()), Details: ErrorDetails{InvalidFilter: InvalidFilter{Field: "corpus", Value: value}}}
}

func suggestions(value string, choices []string) []string {
	value = normalizeToken(value)
	best := ""
	bestDistance := 99
	for _, choice := range choices {
		distance := levenshtein(value, choice)
		if distance < bestDistance {
			bestDistance = distance
			best = choice
		}
	}
	if best != "" && bestDistance <= 3 {
		return []string{best}
	}
	return nil
}

func levenshtein(left, right string) int {
	if left == "" {
		return len(right)
	}
	if right == "" {
		return len(left)
	}
	prev := make([]int, len(right)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(left); i++ {
		current := make([]int, len(right)+1)
		current[0] = i
		for j := 1; j <= len(right); j++ {
			cost := 0
			if left[i-1] != right[j-1] {
				cost = 1
			}
			current[j] = minInt(current[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = current
	}
	return prev[len(right)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}
