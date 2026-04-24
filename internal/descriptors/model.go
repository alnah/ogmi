package descriptors

// SchemaVersion identifies the current descriptor JSON contract version.
const SchemaVersion = "v1"

// Level is a normalized descriptor proficiency level token.
type Level string

// Field names a filterable or groupable descriptor field.
type Field string

// FieldCorpus and related constants name supported descriptor fields.
const (
	FieldCorpus    Field = "corpus"
	FieldDomain    Field = "domain"
	FieldSubdomain Field = "subdomain"
	FieldScale     Field = "scale"
	FieldLevel     Field = "level"
	FieldCode      Field = "code"
	FieldID        Field = "id"
)

// Corpus describes one registered descriptor corpus and its spec layout.
type Corpus struct {
	Name                string   `json:"name"`
	PathFields          []Field  `json:"pathFields"`
	DefaultCoverageAxes []Field  `json:"defaultCoverageAxes"`
	Roots               []string `json:"roots,omitempty"`
	Files               []string `json:"files,omitempty"`
}

// DescriptorRecord is one normalized descriptor row loaded from specs.
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

// DescriptorScaleRecord is one normalized descriptor scale row loaded from specs.
type DescriptorScaleRecord struct {
	Corpus      string   `json:"corpus"`
	Domain      string   `json:"domain,omitempty"`
	Subdomain   string   `json:"subdomain,omitempty"`
	Code        string   `json:"code"`
	ID          string   `json:"id"`
	Description []string `json:"description"`
	File        string   `json:"file"`
}

// Dataset contains loaded descriptor scales and descriptors.
type Dataset struct {
	Scales      []DescriptorScaleRecord `json:"scales"`
	Descriptors []DescriptorRecord      `json:"descriptors"`
}

// LoadOptions selects corpora during descriptor loading.
type LoadOptions struct {
	Corpora []string
}

// Filters selects, sorts, groups, and paginates descriptor records.
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

// QueryResult is the JSON envelope for descriptor list queries.
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

// Group contains descriptor records sharing requested key fields.
type Group struct {
	KeyFields []Field            `json:"keyFields"`
	Key       map[Field]string   `json:"key"`
	Total     int                `json:"total"`
	Items     []DescriptorRecord `json:"items"`
}

// Facet counts descriptor records by one field.
type Facet struct {
	Field   Field         `json:"field"`
	Buckets []FacetBucket `json:"buckets"`
}

// FacetBucket counts one distinct field value.
type FacetBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ScaleQueryResult is the JSON envelope for descriptor scale queries.
type ScaleQueryResult struct {
	Kind          string                  `json:"kind"`
	SchemaVersion string                  `json:"schemaVersion"`
	Total         int                     `json:"total"`
	Returned      int                     `json:"returned"`
	Offset        int                     `json:"offset"`
	Items         []DescriptorScaleRecord `json:"items"`
}

// GetInput identifies one descriptor by corpus plus id or code filters.
type GetInput struct {
	Corpus    string
	ID        string
	Code      string
	Domain    string
	Subdomain string
	Scale     string
	Level     string
}

// GetResult is the JSON envelope for a single descriptor lookup.
type GetResult struct {
	Kind          string           `json:"kind"`
	SchemaVersion string           `json:"schemaVersion"`
	Descriptor    DescriptorRecord `json:"descriptor"`
	Description   []string         `json:"description"`
}

// CompareLevelsInput selects descriptors for level comparison within one scale.
type CompareLevelsInput struct {
	Corpus    string
	Scale     string
	Domain    string
	Subdomain string
	Levels    []string
	Limit     int
	Query     string
}

// LevelSummary reports descriptor presence for one level.
type LevelSummary struct {
	Level   string `json:"level"`
	Total   int    `json:"total"`
	Present bool   `json:"present"`
}

// CompareLevelsResult is the JSON envelope for level comparison output.
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

// CoverageInput selects descriptors for a coverage matrix.
type CoverageInput struct {
	Corpus    string
	Domain    string
	Subdomain string
	Scales    []string
	Levels    []string
}

// CoverageCell counts descriptors for one scale and level pair.
type CoverageCell struct {
	Scale         string   `json:"scale"`
	Level         string   `json:"level"`
	Count         int      `json:"count"`
	DescriptorIDs []string `json:"descriptorIds,omitempty"`
}

// CoverageColumn totals descriptors for one level column.
type CoverageColumn struct {
	Level string `json:"level"`
	Total int    `json:"total"`
}

// CoverageContinuity reports gaps between first and last covered cells.
type CoverageContinuity struct {
	Continuous bool     `json:"continuous"`
	GapColumns []string `json:"gapColumns,omitempty"`
}

// CoverageRow totals coverage for one scale across levels.
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

// CoverageResult is the JSON envelope for descriptor coverage matrices.
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

// SchemaInput selects a schema summary or distinct field values.
type SchemaInput struct {
	Field Field
	Filters
}

// SchemaResult is the JSON envelope for descriptor schema output.
type SchemaResult struct {
	Kind          string   `json:"kind"`
	SchemaVersion string   `json:"schemaVersion"`
	Fields        []Field  `json:"fields,omitempty"`
	Values        []string `json:"values,omitempty"`
}
