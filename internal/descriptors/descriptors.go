package descriptors

import (
	"context"
	"errors"
	"io/fs"
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

func Registry() []Corpus { return nil }

func CanonicalLevels() []string { return nil }

func NormalizeFilters(filters Filters) Filters { return filters }

func Load(ctx context.Context, fsys fs.FS, options LoadOptions) (Dataset, error) {
	_ = ctx
	_ = fsys
	_ = options
	return Dataset{}, CodedError{Code: "internal", Message: "not implemented"}
}

func Query(ctx context.Context, dataset Dataset, filters Filters) (QueryResult, error) {
	_ = ctx
	_ = dataset
	_ = filters
	return QueryResult{}, errors.New("not implemented")
}

func QueryScales(ctx context.Context, dataset Dataset, filters Filters) (ScaleQueryResult, error) {
	_ = ctx
	_ = dataset
	_ = filters
	return ScaleQueryResult{}, errors.New("not implemented")
}

func Get(ctx context.Context, dataset Dataset, input GetInput) (GetResult, error) {
	_ = ctx
	_ = dataset
	_ = input
	return GetResult{}, errors.New("not implemented")
}

func CompareLevels(ctx context.Context, dataset Dataset, input CompareLevelsInput) (CompareLevelsResult, error) {
	_ = ctx
	_ = dataset
	_ = input
	return CompareLevelsResult{}, errors.New("not implemented")
}

func Coverage(ctx context.Context, dataset Dataset, input CoverageInput) (CoverageResult, error) {
	_ = ctx
	_ = dataset
	_ = input
	return CoverageResult{}, errors.New("not implemented")
}

func Schema(ctx context.Context, dataset Dataset, input SchemaInput) (SchemaResult, error) {
	_ = ctx
	_ = dataset
	_ = input
	return SchemaResult{}, errors.New("not implemented")
}
