package descriptors

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
