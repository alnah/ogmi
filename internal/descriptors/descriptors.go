package descriptors

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "v1"

type Level string

type Field string

const (
	FieldCorpus    Field = "corpus"
	FieldDomain    Field = "domain"
	FieldSubdomain Field = "subdomain"
	FieldScale     Field = "scale"
	FieldLevel     Field = "level"
	FieldCode      Field = "code"
	FieldID        Field = "id"
)

type Corpus struct {
	Name                string   `json:"name"`
	PathFields          []Field  `json:"pathFields"`
	DefaultCoverageAxes []Field  `json:"defaultCoverageAxes"`
	Roots               []string `json:"roots,omitempty"`
	Files               []string `json:"files,omitempty"`
}

type DescriptorRecord struct {
	Corpus      string `json:"corpus"`
	Domain      string `json:"domain,omitempty"`
	Subdomain   string `json:"subdomain,omitempty"`
	Scale       string `json:"scale"`
	Level       string `json:"level"`
	Code        string `json:"code"`
	ID          string `json:"id"`
	Description string `json:"description"`
	File        string `json:"file"`
}

type DescriptorScaleRecord struct {
	Corpus      string   `json:"corpus"`
	Domain      string   `json:"domain,omitempty"`
	Subdomain   string   `json:"subdomain,omitempty"`
	Code        string   `json:"code"`
	ID          string   `json:"id"`
	Description []string `json:"description"`
	File        string   `json:"file"`
}

type Dataset struct {
	Scales      []DescriptorScaleRecord `json:"scales"`
	Descriptors []DescriptorRecord      `json:"descriptors"`
}

type LoadOptions struct {
	Corpora []string
}

type Filters struct {
	Corpora    []string
	Domain     string
	Subdomain  string
	Scales     []string
	Levels     []string
	Code       string
	ID         string
	Query      string
	Offset     int
	Limit      int
	SortBy     Field
	SortOrder  string
	GroupBy    []Field
	WithFacets bool
}

type QueryResult struct {
	Kind          string             `json:"kind"`
	SchemaVersion string             `json:"schemaVersion"`
	Filters       Filters            `json:"filters"`
	Total         int                `json:"total"`
	Returned      int                `json:"returned"`
	Offset        int                `json:"offset"`
	Items         []DescriptorRecord `json:"items"`
	Groups        []Group            `json:"groups,omitempty"`
	Facets        []Facet            `json:"facets,omitempty"`
}

type Group struct {
	KeyFields []Field            `json:"keyFields"`
	Key       map[Field]string   `json:"key"`
	Total     int                `json:"total"`
	Items     []DescriptorRecord `json:"items"`
}

type Facet struct {
	Field   Field         `json:"field"`
	Buckets []FacetBucket `json:"buckets"`
}

type FacetBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type ScaleQueryResult struct {
	Kind          string                  `json:"kind"`
	SchemaVersion string                  `json:"schemaVersion"`
	Total         int                     `json:"total"`
	Returned      int                     `json:"returned"`
	Offset        int                     `json:"offset"`
	Items         []DescriptorScaleRecord `json:"items"`
}

type GetInput struct {
	Corpus    string
	ID        string
	Code      string
	Domain    string
	Subdomain string
	Scale     string
	Level     string
}

type GetResult struct {
	Kind          string           `json:"kind"`
	SchemaVersion string           `json:"schemaVersion"`
	Descriptor    DescriptorRecord `json:"descriptor"`
	Description   []string         `json:"description"`
}

type CompareLevelsInput struct {
	Corpus    string
	Scale     string
	Domain    string
	Subdomain string
	Levels    []string
	Limit     int
	Query     string
}

type LevelSummary struct {
	Level   string `json:"level"`
	Total   int    `json:"total"`
	Present bool   `json:"present"`
}

type CompareLevelsResult struct {
	Kind          string             `json:"kind"`
	SchemaVersion string             `json:"schemaVersion"`
	Levels        []string           `json:"levels"`
	Summaries     []LevelSummary     `json:"summaries"`
	MissingLevels []string           `json:"missingLevels"`
	Total         int                `json:"total"`
	Returned      int                `json:"returned"`
	Items         []DescriptorRecord `json:"items"`
}

type CoverageInput struct {
	Corpus    string
	Domain    string
	Subdomain string
	Scales    []string
	Levels    []string
}

type CoverageCell struct {
	Scale         string   `json:"scale"`
	Level         string   `json:"level"`
	Count         int      `json:"count"`
	DescriptorIDs []string `json:"descriptorIds,omitempty"`
}

type CoverageColumn struct {
	Level string `json:"level"`
	Total int    `json:"total"`
}

type CoverageContinuity struct {
	Continuous bool     `json:"continuous"`
	GapColumns []string `json:"gapColumns,omitempty"`
}

type CoverageRow struct {
	Scale              string             `json:"scale"`
	Total              int                `json:"total"`
	CoveredColumns     []string           `json:"coveredColumns"`
	MissingColumns     []string           `json:"missingColumns"`
	FirstCoveredColumn string             `json:"firstCoveredColumn,omitempty"`
	LastCoveredColumn  string             `json:"lastCoveredColumn,omitempty"`
	Continuity         CoverageContinuity `json:"continuity"`
	Cells              []CoverageCell     `json:"cells"`
}

type CoverageResult struct {
	Kind          string           `json:"kind"`
	SchemaVersion string           `json:"schemaVersion"`
	Levels        []string         `json:"levels"`
	Scales        []string         `json:"scales"`
	Columns       []CoverageColumn `json:"columns"`
	Rows          []CoverageRow    `json:"rows"`
	Cells         []CoverageCell   `json:"cells"`
	GrandTotal    int              `json:"grandTotal"`
}

type SchemaInput struct {
	Field Field
	Filters
}

type SchemaResult struct {
	Kind          string   `json:"kind"`
	SchemaVersion string   `json:"schemaVersion"`
	Fields        []Field  `json:"fields,omitempty"`
	Values        []string `json:"values,omitempty"`
}

type InvalidFilter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type ErrorDetails struct {
	InvalidFilter   InvalidFilter `json:"invalidFilter,omitempty"`
	AvailableFields []Field       `json:"availableFields,omitempty"`
}

type CodedError struct {
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	Suggestions []string     `json:"suggestions,omitempty"`
	Details     ErrorDetails `json:"details,omitempty"`
}

func (e CodedError) Error() string { return e.Message }

var canonicalLevels = []string{"pre_a1", "a1", "a2", "b1", "b2", "c1", "c2"}

var registry = []Corpus{
	{Name: "cefr", Roots: []string{"specs/cefr"}, PathFields: []Field{FieldDomain, FieldSubdomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldSubdomain, FieldLevel}},
	{Name: "french", Roots: []string{"specs/french"}, PathFields: []Field{FieldDomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldLevel}},
	{Name: "texts", Roots: []string{"specs/texts"}, PathFields: []Field{FieldDomain}, DefaultCoverageAxes: []Field{FieldCorpus, FieldDomain, FieldLevel}},
	{Name: "themes", Files: []string{"specs/themes/descriptors.yml"}, PathFields: []Field{}, DefaultCoverageAxes: []Field{FieldCorpus, FieldLevel}},
}

func Registry() []Corpus {
	out := make([]Corpus, len(registry))
	copy(out, registry)
	return out
}

func CanonicalLevels() []string {
	out := make([]string, len(canonicalLevels))
	copy(out, canonicalLevels)
	return out
}

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

func Load(ctx context.Context, fsys fs.FS, options LoadOptions) (Dataset, error) {
	if err := ctx.Err(); err != nil {
		return Dataset{}, CodedError{Code: "internal", Message: err.Error()}
	}
	corpora, err := resolveCorpora(options.Corpora)
	if err != nil {
		return Dataset{}, err
	}

	var dataset Dataset
	for _, corpus := range corpora {
		files := descriptorFiles(fsys, corpus)
		for _, file := range files {
			part, err := loadFile(fsys, corpus, file)
			if err != nil {
				return Dataset{}, err
			}
			dataset.Scales = append(dataset.Scales, part.Scales...)
			dataset.Descriptors = append(dataset.Descriptors, part.Descriptors...)
		}
	}
	if len(dataset.Scales) == 0 && len(dataset.Descriptors) == 0 {
		return Dataset{}, CodedError{Code: "missing_specs", Message: "Descriptor specs not found"}
	}
	sort.SliceStable(dataset.Scales, func(i, j int) bool { return compareScale(dataset.Scales[i], dataset.Scales[j]) < 0 })
	sort.SliceStable(dataset.Descriptors, func(i, j int) bool { return compareDescriptor(dataset.Descriptors[i], dataset.Descriptors[j]) < 0 })
	return dataset, nil
}

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

func Coverage(ctx context.Context, dataset Dataset, input CoverageInput) (CoverageResult, error) {
	if err := ctx.Err(); err != nil {
		return CoverageResult{}, CodedError{Code: "internal", Message: err.Error()}
	}
	corpus := normalizeToken(input.Corpus)
	if corpus == "" {
		return CoverageResult{}, CodedError{Code: "missing_required_filter", Message: "coverage requires corpus"}
	}
	if !knownCorpus(corpus) {
		return CoverageResult{}, unknownCorpusError(corpus)
	}
	filters := NormalizeFilters(Filters{Corpora: []string{corpus}, Domain: input.Domain, Subdomain: input.Subdomain, Scales: input.Scales, Levels: input.Levels})
	items := filterDescriptors(dataset.Descriptors, filters)
	sort.SliceStable(items, func(i, j int) bool { return compareDescriptor(items[i], items[j]) < 0 })
	levels := filters.Levels
	if len(levels) == 0 {
		levels = CanonicalLevels()
	}
	scales := filters.Scales
	if len(scales) == 0 {
		scales = scalesFromItems(items)
	}
	columns := make([]CoverageColumn, 0, len(levels))
	rows := make([]CoverageRow, 0, len(scales))
	flatCells := make([]CoverageCell, 0, len(scales)*len(levels))
	grand := 0
	for _, level := range levels {
		columns = append(columns, CoverageColumn{Level: level})
	}
	for _, scale := range scales {
		row := CoverageRow{Scale: scale, CoveredColumns: []string{}, MissingColumns: []string{}, Cells: []CoverageCell{}}
		for colIndex, level := range levels {
			ids := idsForCell(items, scale, level)
			cell := CoverageCell{Scale: scale, Level: level, Count: len(ids), DescriptorIDs: ids}
			row.Total += cell.Count
			columns[colIndex].Total += cell.Count
			grand += cell.Count
			if cell.Count > 0 {
				row.CoveredColumns = append(row.CoveredColumns, level)
				if row.FirstCoveredColumn == "" {
					row.FirstCoveredColumn = level
				}
				row.LastCoveredColumn = level
			} else {
				row.MissingColumns = append(row.MissingColumns, level)
			}
			row.Cells = append(row.Cells, cell)
			flatCells = append(flatCells, cell)
		}
		row.Continuity = coverageContinuity(row.Cells)
		rows = append(rows, row)
	}
	return CoverageResult{Kind: "descriptor_coverage_matrix", SchemaVersion: SchemaVersion, Levels: levels, Scales: scales, Columns: columns, Rows: rows, Cells: flatCells, GrandTotal: grand}, nil
}

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

type yamlFile struct {
	Catalog []yamlRow `yaml:"catalog"`
	Entries []yamlRow `yaml:"entries"`
}

type yamlRow struct {
	Scale       string      `yaml:"scale"`
	Level       string      `yaml:"level"`
	Code        string      `yaml:"code"`
	ID          string      `yaml:"id"`
	Description description `yaml:"description"`
}

type description []string

func (d *description) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var text string
		if err := value.Decode(&text); err != nil {
			return err
		}
		*d = []string{text}
		return nil
	case yaml.SequenceNode:
		var lines []string
		if err := value.Decode(&lines); err != nil {
			return err
		}
		*d = lines
		return nil
	default:
		return fmt.Errorf("description must be string or list")
	}
}

func descriptorFiles(fsys fs.FS, corpus Corpus) []string {
	seen := make(map[string]bool)
	files := []string{}
	for _, root := range corpus.Roots {
		_ = fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !entry.IsDir() && path.Base(name) == "descriptors.yml" && !seen[name] {
				seen[name] = true
				files = append(files, name)
			}
			return nil
		})
	}
	for _, file := range corpus.Files {
		if _, err := fs.Stat(fsys, file); err == nil && !seen[file] {
			seen[file] = true
			files = append(files, file)
		}
	}
	sort.Strings(files)
	return files
}

func loadFile(fsys fs.FS, corpus Corpus, file string) (Dataset, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return Dataset{}, CodedError{Code: "missing_specs", Message: fmt.Sprintf("Descriptor spec %s not found", file)}
	}
	var raw yamlFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Dataset{}, CodedError{Code: "invalid_yaml", Message: fmt.Sprintf("Invalid descriptor YAML in %s: %v", file, err)}
	}
	values := pathValues(corpus, file)
	var dataset Dataset
	for index, row := range raw.Catalog {
		code := normalizeToken(row.Code)
		id := strings.TrimSpace(row.ID)
		desc := normalizeDescription([]string(row.Description))
		if code == "" || id == "" || len(desc) == 0 {
			return Dataset{}, CodedError{Code: "invalid_row", Message: fmt.Sprintf("Invalid descriptor catalog row %d in %s", index+1, file)}
		}
		dataset.Scales = append(dataset.Scales, DescriptorScaleRecord{Corpus: corpus.Name, Domain: values.Domain, Subdomain: values.Subdomain, Code: code, ID: id, Description: desc, File: file})
	}
	for index, row := range raw.Entries {
		scale := normalizeToken(row.Scale)
		level := normalizeLevel(row.Level)
		code := normalizeToken(row.Code)
		id := strings.TrimSpace(row.ID)
		desc := normalizeDescription([]string(row.Description))
		if scale == "" || level == "" || code == "" || id == "" || len(desc) == 0 {
			return Dataset{}, CodedError{Code: "invalid_row", Message: fmt.Sprintf("Invalid descriptor entries row %d in %s", index+1, file)}
		}
		dataset.Descriptors = append(dataset.Descriptors, DescriptorRecord{Corpus: corpus.Name, Domain: values.Domain, Subdomain: values.Subdomain, Scale: scale, Level: level, Code: code, ID: id, Description: strings.Join(desc, "\n"), File: file})
	}
	return dataset, nil
}

type pathFieldValues struct{ Domain, Subdomain string }

func pathValues(corpus Corpus, file string) pathFieldValues {
	for _, root := range corpus.Roots {
		rel, ok := strings.CutPrefix(file, strings.TrimSuffix(root, "/")+"/")
		if !ok {
			continue
		}
		segments := strings.Split(path.Dir(rel), "/")
		if path.Dir(rel) == "." {
			segments = nil
		}
		values := pathFieldValues{}
		if len(corpus.PathFields) > 0 && len(segments) > 0 && corpus.PathFields[0] == FieldDomain {
			values.Domain = normalizeToken(segments[0])
		}
		if len(corpus.PathFields) > 1 && len(segments) > 1 && corpus.PathFields[1] == FieldSubdomain {
			values.Subdomain = normalizeToken(segments[1])
		}
		return values
	}
	return pathFieldValues{}
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

func unknownCorpusError(value string) CodedError {
	return CodedError{Code: "unknown_corpus", Message: fmt.Sprintf("Unknown descriptor corpus: %s", value), Suggestions: suggestions(value, corpusNames()), Details: ErrorDetails{InvalidFilter: InvalidFilter{Field: "corpus", Value: value}}}
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

func compareDescriptorForQuery(left, right DescriptorRecord, field Field, order string) int {
	if field == "" {
		return compareDescriptor(left, right)
	}
	base := compareRecordField(left, right, field)
	if base != 0 && order == "desc" {
		return -base
	}
	if base != 0 {
		return base
	}
	return compareDescriptor(left, right)
}

func compareScaleForQuery(left, right DescriptorScaleRecord, field Field, order string) int {
	if field == "" {
		return compareScale(left, right)
	}
	base := strings.Compare(scaleField(left, field), scaleField(right, field))
	if base != 0 && order == "desc" {
		return -base
	}
	if base != 0 {
		return base
	}
	return compareScale(left, right)
}

func compareDescriptor(left, right DescriptorRecord) int {
	for _, diff := range []int{strings.Compare(left.Corpus, right.Corpus), strings.Compare(left.Domain, right.Domain), strings.Compare(left.Subdomain, right.Subdomain), strings.Compare(left.Scale, right.Scale), compareLevels(left.Level, right.Level), strings.Compare(left.Code, right.Code), strings.Compare(left.ID, right.ID)} {
		if diff != 0 {
			return diff
		}
	}
	return 0
}

func compareScale(left, right DescriptorScaleRecord) int {
	for _, diff := range []int{strings.Compare(left.Corpus, right.Corpus), strings.Compare(left.Domain, right.Domain), strings.Compare(left.Subdomain, right.Subdomain), strings.Compare(left.Code, right.Code), strings.Compare(left.ID, right.ID)} {
		if diff != 0 {
			return diff
		}
	}
	return 0
}

func compareRecordField(left, right DescriptorRecord, field Field) int {
	if field == FieldLevel {
		return compareLevels(left.Level, right.Level)
	}
	return strings.Compare(recordField(left, field), recordField(right, field))
}

func recordField(record DescriptorRecord, field Field) string {
	switch field {
	case FieldCorpus:
		return record.Corpus
	case FieldDomain:
		return record.Domain
	case FieldSubdomain:
		return record.Subdomain
	case FieldScale:
		return record.Scale
	case FieldLevel:
		return record.Level
	case FieldCode:
		return record.Code
	case FieldID:
		return record.ID
	default:
		return ""
	}
}

func scaleField(record DescriptorScaleRecord, field Field) string {
	switch field {
	case FieldCorpus:
		return record.Corpus
	case FieldDomain:
		return record.Domain
	case FieldSubdomain:
		return record.Subdomain
	case FieldScale, FieldCode:
		return record.Code
	case FieldID:
		return record.ID
	default:
		return ""
	}
}

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

func buildGroups(items []DescriptorRecord, fields []Field) []Group {
	if len(fields) == 0 {
		return nil
	}
	groups := []Group{}
	indexes := make(map[string]int)
	for _, item := range items {
		key := make(map[Field]string)
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			value := recordField(item, field)
			if value != "" {
				key[field] = value
			}
			parts = append(parts, string(field)+"="+value)
		}
		id := strings.Join(parts, "\x00")
		index, ok := indexes[id]
		if !ok {
			indexes[id] = len(groups)
			groups = append(groups, Group{KeyFields: fields, Key: key, Total: 1, Items: []DescriptorRecord{item}})
			continue
		}
		groups[index].Total++
		groups[index].Items = append(groups[index].Items, item)
	}
	return groups
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

func scalesFromItems(items []DescriptorRecord) []string {
	seen := make(map[string]bool)
	scales := []string{}
	for _, item := range items {
		if !seen[item.Scale] {
			seen[item.Scale] = true
			scales = append(scales, item.Scale)
		}
	}
	sort.Strings(scales)
	return scales
}

func idsForCell(items []DescriptorRecord, scale, level string) []string {
	var ids []string
	for _, item := range items {
		if item.Scale == scale && item.Level == level {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func coverageContinuity(cells []CoverageCell) CoverageContinuity {
	first := -1
	last := -1
	for index, cell := range cells {
		if cell.Count > 0 {
			if first == -1 {
				first = index
			}
			last = index
		}
	}
	if first == -1 {
		return CoverageContinuity{Continuous: false}
	}
	gaps := []string{}
	for index := first; index <= last; index++ {
		if cells[index].Count == 0 {
			gaps = append(gaps, cells[index].Level)
		}
	}
	if len(gaps) == 0 {
		return CoverageContinuity{Continuous: true}
	}
	return CoverageContinuity{Continuous: false, GapColumns: gaps}
}

func splitDescription(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
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

func corpusNames() []string {
	names := make([]string, 0, len(registry))
	for _, corpus := range registry {
		names = append(names, corpus.Name)
	}
	sort.Strings(names)
	return names
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

func compareLevels(left, right string) int {
	li, lok := levelIndex(left)
	ri, rok := levelIndex(right)
	if lok && rok {
		return li - ri
	}
	if lok {
		return -1
	}
	if rok {
		return 1
	}
	return strings.Compare(left, right)
}

func levelIndex(level string) (int, bool) {
	for index, candidate := range canonicalLevels {
		if candidate == level {
			return index, true
		}
	}
	return 0, false
}

func isCanonicalLevel(level string) bool { _, ok := levelIndex(level); return ok }

func stringIn(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func fieldIn(value Field, values []Field) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
