package descriptors

import "sort"

var registry = []Corpus{
	{Name: "cefr", Roots: []string{"specs/cefr"}, PathFields: []Field{FieldDomain, FieldSubdomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldSubdomain, FieldLevel}},
	{Name: "french", Roots: []string{"specs/french"}, PathFields: []Field{FieldDomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldLevel}},
	{Name: "texts", Roots: []string{"specs/texts"}, PathFields: []Field{FieldDomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldLevel}},
	{Name: "themes", Files: []string{"specs/themes/descriptors.yml"}, PathFields: []Field{}, DefaultCoverageAxes: []Field{FieldCorpus, FieldLevel}},
}

// Registry returns a copy of registered descriptor corpora.
func Registry() []Corpus {
	out := make([]Corpus, len(registry))
	copy(out, registry)
	return out
}

// CanonicalLevels returns CEFR levels in canonical order.
func CanonicalLevels() []string {
	out := make([]string, len(canonicalLevels))
	copy(out, canonicalLevels)
	return out
}

func resolveCorpora(requested []string) ([]Corpus, error) {
	names := normalizeTokens(requested)
	if len(names) == 0 {
		out := Registry()
		sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
		return out, nil
	}
	out := make([]Corpus, 0, len(names))
	for _, name := range names {
		corpus, ok := corpusByName(name)
		if !ok {
			return nil, unknownCorpusError(name)
		}
		out = append(out, corpus)
	}
	return out, nil
}

func corpusByName(name string) (Corpus, bool) {
	for _, corpus := range registry {
		if corpus.Name == name {
			return corpus, true
		}
	}
	return Corpus{}, false
}

func knownCorpus(name string) bool { _, ok := corpusByName(name); return ok }

func corpusNames() []string {
	names := make([]string, 0, len(registry))
	for _, corpus := range registry {
		names = append(names, corpus.Name)
	}
	sort.Strings(names)
	return names
}
