package descriptors_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func queryDataset() descriptors.Dataset {
	return descriptors.Dataset{
		Scales: []descriptors.DescriptorScaleRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "addressing_audiences", ID: "cefr.production.speaking.descriptors.addressing_audiences", Description: []string{"Address audiences."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "turntaking", ID: "cefr.production.speaking.descriptors.turntaking", Description: []string{"Manage turntaking."}, File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Code: "personal_identity", ID: "themes.descriptors.personal_identity", Description: []string{"Personal identity themes."}, File: "specs/themes/descriptors.yml"},
		},
		Descriptors: []descriptors.DescriptorRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a1", Code: "deliver_toast", ID: "cefr.production.speaking.descriptors.addressing_audiences.deliver_toast.a1", Description: "Can deliver a short toast.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "addressing_audiences", Level: "a2", Code: "present_simple_announcement", ID: "cefr.production.speaking.descriptors.addressing_audiences.present_simple_announcement.a2", Description: "Can present a simple announcement.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "a1", Code: "open_simple_exchange", ID: "cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1", Description: "Can open a simple exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "b1", Code: "maintain_exchange", ID: "cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1", Description: "Can maintain an exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
			{Corpus: "themes", Scale: "personal_identity", Level: "a1", Code: "share_basic_identity_facts", ID: "themes.descriptors.personal_identity.share_basic_identity_facts.a1", Description: "Can share basic identity facts.", File: "specs/themes/descriptors.yml"},
		},
	}
}

type descriptorGroupSummary struct {
	KeyFields []descriptors.Field
	Scale     string
	Level     string
	Total     int
	ItemIDs   []string
}

func descriptorIDs(records []descriptors.DescriptorRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	return ids
}

func groupSummaries(groups []descriptors.Group) []descriptorGroupSummary {
	summaries := make([]descriptorGroupSummary, 0, len(groups))
	for _, group := range groups {
		summaries = append(summaries, descriptorGroupSummary{
			KeyFields: group.KeyFields,
			Scale:     group.Key[descriptors.FieldScale],
			Level:     group.Key[descriptors.FieldLevel],
			Total:     group.Total,
			ItemIDs:   descriptorIDs(group.Items),
		})
	}
	return summaries
}

func facetBuckets(facets []descriptors.Facet, field descriptors.Field) []descriptors.FacetBucket {
	for _, facet := range facets {
		if facet.Field == field {
			return facet.Buckets
		}
	}
	return nil
}

func coverageRow(rows []descriptors.CoverageRow, scale string) (descriptors.CoverageRow, bool) {
	for _, row := range rows {
		if row.Scale == scale {
			return row, true
		}
	}
	return descriptors.CoverageRow{}, false
}

func coverageCell(cells []descriptors.CoverageCell, level string) (descriptors.CoverageCell, bool) {
	for _, cell := range cells {
		if cell.Level == level {
			return cell, true
		}
	}
	return descriptors.CoverageCell{}, false
}

func fieldsContain(fields []descriptors.Field, want descriptors.Field) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}

func requireCodedError(t *testing.T, err error, wantCode string) descriptors.CodedError {
	t.Helper()
	var coded descriptors.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("error = %T %[1]v, want descriptors.CodedError", err)
	}
	if coded.Code != wantCode {
		t.Fatalf("coded error code = %q, want %q; error = %+v", coded.Code, wantCode, coded)
	}
	return coded
}

func TestQueryFiltersSortsPaginatesGroupsAndFacets(t *testing.T) {
	got, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{
		Corpora:    []string{"CEFR"},
		Domain:     "Production",
		Subdomain:  "Speaking",
		Scales:     []string{"turntaking", "addressing_audiences"},
		Levels:     []string{"b1", "a1"},
		Offset:     1,
		Limit:      2,
		SortBy:     descriptors.FieldCode,
		SortOrder:  "asc",
		GroupBy:    []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel},
		WithFacets: true,
	})
	if err != nil {
		t.Fatalf("descriptors.Query() error = %v", err)
	}

	wantItems := []descriptors.DescriptorRecord{
		{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "b1", Code: "maintain_exchange", ID: "cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1", Description: "Can maintain an exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
		{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Scale: "turntaking", Level: "a1", Code: "open_simple_exchange", ID: "cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1", Description: "Can open a simple exchange.", File: "specs/cefr/production/speaking/descriptors.yml"},
	}
	if got.Kind != "descriptor_list" || got.SchemaVersion != descriptors.SchemaVersion || got.Total != 3 || got.Returned != 2 || got.Offset != 1 {
		t.Errorf("descriptors.Query() metadata = kind %q schema %q total %d returned %d offset %d, want descriptor_list %s 3 2 1", got.Kind, got.SchemaVersion, got.Total, got.Returned, got.Offset, descriptors.SchemaVersion)
	}
	if diff := cmp.Diff(wantItems, got.Items); diff != "" {
		t.Errorf("descriptors.Query() items mismatch (-want +got):\n%s", diff)
	}

	wantScaleBuckets := []descriptors.FacetBucket{{Value: "addressing_audiences", Count: 1}, {Value: "turntaking", Count: 2}}
	if diff := cmp.Diff(wantScaleBuckets, facetBuckets(got.Facets, descriptors.FieldScale)); diff != "" {
		t.Errorf("descriptors.Query() scale facet buckets mismatch (-want +got):\n%s", diff)
	}
	wantLevelBuckets := []descriptors.FacetBucket{{Value: "a1", Count: 2}, {Value: "b1", Count: 1}}
	if diff := cmp.Diff(wantLevelBuckets, facetBuckets(got.Facets, descriptors.FieldLevel)); diff != "" {
		t.Errorf("descriptors.Query() level facet buckets mismatch (-want +got):\n%s", diff)
	}

	wantGroups := []descriptorGroupSummary{
		{KeyFields: []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel}, Scale: "turntaking", Level: "b1", Total: 1, ItemIDs: []string{"cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1"}},
		{KeyFields: []descriptors.Field{descriptors.FieldScale, descriptors.FieldLevel}, Scale: "turntaking", Level: "a1", Total: 1, ItemIDs: []string{"cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1"}},
	}
	if diff := cmp.Diff(wantGroups, groupSummaries(got.Groups)); diff != "" {
		t.Errorf("descriptors.Query() groups mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryReportsTypedInvalidFilterDetails(t *testing.T) {
	_, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Levels: []string{"Z9"}})
	coded := requireCodedError(t, err, "invalid_filter")
	wantInvalidFilter := descriptors.InvalidFilter{Field: "level", Value: "Z9"}
	if diff := cmp.Diff(wantInvalidFilter, coded.Details.InvalidFilter); diff != "" {
		t.Errorf("invalidFilter mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryScalesRequiresCorpusAndReturnsScaleRecords(t *testing.T) {
	got, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefr"}, Domain: "production", Query: "audiences"})
	if err != nil {
		t.Fatalf("descriptors.QueryScales() error = %v", err)
	}
	want := descriptors.ScaleQueryResult{
		Kind:          "descriptor_scales",
		SchemaVersion: descriptors.SchemaVersion,
		Total:         1,
		Returned:      1,
		Offset:        0,
		Items: []descriptors.DescriptorScaleRecord{
			{Corpus: "cefr", Domain: "production", Subdomain: "speaking", Code: "addressing_audiences", ID: "cefr.production.speaking.descriptors.addressing_audiences", Description: []string{"Address audiences."}, File: "specs/cefr/production/speaking/descriptors.yml"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("descriptors.QueryScales() mismatch (-want +got):\n%s", diff)
	}
}

func TestDescriptorSadPathsReturnStableCodedErrors(t *testing.T) {
	tests := []struct {
		name     string
		run      func() error
		wantCode string
	}{
		{name: "scales without corpus", run: func() error {
			_, err := descriptors.QueryScales(context.Background(), queryDataset(), descriptors.Filters{})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "get without id or code", run: func() error {
			_, err := descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{Corpus: "cefr"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "compare without corpus", run: func() error {
			_, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Scale: "turntaking"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "compare without scale", run: func() error {
			_, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr"})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "coverage without corpus", run: func() error {
			_, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{})
			return err
		}, wantCode: "missing_required_filter"},
		{name: "unknown corpus", run: func() error {
			_, err := descriptors.Query(context.Background(), queryDataset(), descriptors.Filters{Corpora: []string{"cefrr"}})
			return err
		}, wantCode: "unknown_corpus"},
		{name: "missing requested specs", run: func() error {
			_, err := descriptors.Load(context.Background(), fstest.MapFS{}, descriptors.LoadOptions{Corpora: []string{"themes"}})
			return err
		}, wantCode: "missing_specs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = requireCodedError(t, tt.run(), tt.wantCode)
		})
	}
}

func TestGetPrefersIDAndReportsAmbiguousCode(t *testing.T) {
	got, err := descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{Corpus: "cefr", Code: "deliver_toast", ID: "cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1"})
	if err != nil {
		t.Fatalf("descriptors.Get(id wins) error = %v", err)
	}
	if got.Kind != "descriptor_get" || got.Descriptor.Code != "open_simple_exchange" {
		t.Errorf("descriptors.Get(id wins) = %+v, want open_simple_exchange descriptor_get", got)
	}

	_, err = descriptors.Get(context.Background(), queryDataset(), descriptors.GetInput{Corpus: "cefr", Code: "open_simple_exchange"})
	if err != nil {
		t.Fatalf("descriptors.Get(unique code) error = %v", err)
	}

	ambiguousDataset := queryDataset()
	ambiguousDataset.Descriptors = append(ambiguousDataset.Descriptors, descriptors.DescriptorRecord{Corpus: "cefr", Domain: "interaction", Subdomain: "speaking", Scale: "turntaking", Level: "a2", Code: "open_simple_exchange", ID: "cefr.interaction.speaking.descriptors.turntaking.open_simple_exchange.a2", Description: "Can open another exchange.", File: "specs/cefr/interaction/speaking/descriptors.yml"})
	_, err = descriptors.Get(context.Background(), ambiguousDataset, descriptors.GetInput{Corpus: "cefr", Code: "open_simple_exchange"})
	_ = requireCodedError(t, err, "ambiguous_lookup")
}

func TestCompareLevelsReturnsCompleteSummariesBeforeLimit(t *testing.T) {
	got, err := descriptors.CompareLevels(context.Background(), queryDataset(), descriptors.CompareLevelsInput{Corpus: "cefr", Scale: "turntaking", Levels: []string{"a1", "a2", "b1"}, Limit: 1})
	if err != nil {
		t.Fatalf("descriptors.CompareLevels() error = %v", err)
	}

	if got.Kind != "descriptor_level_comparison" || got.SchemaVersion != descriptors.SchemaVersion || got.Total != 2 || got.Returned != 1 {
		t.Errorf("descriptors.CompareLevels() metadata = kind %q schema %q total %d returned %d, want descriptor_level_comparison %s 2 1", got.Kind, got.SchemaVersion, got.Total, got.Returned, descriptors.SchemaVersion)
	}
	wantLevels := []string{"a1", "a2", "b1"}
	if diff := cmp.Diff(wantLevels, got.Levels); diff != "" {
		t.Errorf("descriptors.CompareLevels() levels mismatch (-want +got):\n%s", diff)
	}
	wantSummaries := []descriptors.LevelSummary{{Level: "a1", Total: 1, Present: true}, {Level: "a2", Total: 0, Present: false}, {Level: "b1", Total: 1, Present: true}}
	if diff := cmp.Diff(wantSummaries, got.Summaries); diff != "" {
		t.Errorf("descriptors.CompareLevels() summaries mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"a2"}, got.MissingLevels); diff != "" {
		t.Errorf("descriptors.CompareLevels() missing levels mismatch (-want +got):\n%s", diff)
	}
	if gotItems := len(got.Items); gotItems != 1 {
		t.Errorf("descriptors.CompareLevels() items len = %d, want limit to narrow returned items only", gotItems)
	}
}

func TestCoverageMatrixReportsRichRowsColumnsCellsAndContinuity(t *testing.T) {
	got, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{Corpus: "cefr", Domain: "production", Scales: []string{"addressing_audiences", "turntaking"}, Levels: []string{"a1", "a2", "b1"}})
	if err != nil {
		t.Fatalf("descriptors.Coverage() error = %v", err)
	}
	if got.Kind != "descriptor_coverage_matrix" || got.SchemaVersion != descriptors.SchemaVersion {
		t.Errorf("descriptors.Coverage() kind/schema = %q/%q", got.Kind, got.SchemaVersion)
	}
	if diff := cmp.Diff([]string{"addressing_audiences", "turntaking"}, got.Scales); diff != "" {
		t.Errorf("descriptors.Coverage() scales mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"a1", "a2", "b1"}, got.Levels); diff != "" {
		t.Errorf("descriptors.Coverage() levels mismatch (-want +got):\n%s", diff)
	}

	wantColumns := []descriptors.CoverageColumn{{Level: "a1", Total: 2}, {Level: "a2", Total: 1}, {Level: "b1", Total: 1}}
	if diff := cmp.Diff(wantColumns, got.Columns); diff != "" {
		t.Errorf("descriptors.Coverage() columns mismatch (-want +got):\n%s", diff)
	}
	wantRows := []descriptors.CoverageRow{
		{
			Scale:              "addressing_audiences",
			Total:              2,
			CoveredColumns:     []string{"a1", "a2"},
			MissingColumns:     []string{"b1"},
			FirstCoveredColumn: "a1",
			LastCoveredColumn:  "a2",
			Continuity:         descriptors.CoverageContinuity{Continuous: true},
			Cells: []descriptors.CoverageCell{
				{Scale: "addressing_audiences", Level: "a1", Count: 1, DescriptorIDs: []string{"cefr.production.speaking.descriptors.addressing_audiences.deliver_toast.a1"}},
				{Scale: "addressing_audiences", Level: "a2", Count: 1, DescriptorIDs: []string{"cefr.production.speaking.descriptors.addressing_audiences.present_simple_announcement.a2"}},
				{Scale: "addressing_audiences", Level: "b1", Count: 0, DescriptorIDs: nil},
			},
		},
		{
			Scale:              "turntaking",
			Total:              2,
			CoveredColumns:     []string{"a1", "b1"},
			MissingColumns:     []string{"a2"},
			FirstCoveredColumn: "a1",
			LastCoveredColumn:  "b1",
			Continuity:         descriptors.CoverageContinuity{Continuous: false, GapColumns: []string{"a2"}},
			Cells: []descriptors.CoverageCell{
				{Scale: "turntaking", Level: "a1", Count: 1, DescriptorIDs: []string{"cefr.production.speaking.descriptors.turntaking.open_simple_exchange.a1"}},
				{Scale: "turntaking", Level: "a2", Count: 0, DescriptorIDs: nil},
				{Scale: "turntaking", Level: "b1", Count: 1, DescriptorIDs: []string{"cefr.production.speaking.descriptors.turntaking.maintain_exchange.b1"}},
			},
		},
	}
	if diff := cmp.Diff(wantRows, got.Rows); diff != "" {
		t.Errorf("descriptors.Coverage() rows mismatch (-want +got):\n%s", diff)
	}
	wantCells := append([]descriptors.CoverageCell{}, wantRows[0].Cells...)
	wantCells = append(wantCells, wantRows[1].Cells...)
	if diff := cmp.Diff(wantCells, got.Cells); diff != "" {
		t.Errorf("descriptors.Coverage() flat cells mismatch (-want +got):\n%s", diff)
	}
	if got.GrandTotal != 4 {
		t.Errorf("descriptors.Coverage() grand total = %d, want 4", got.GrandTotal)
	}
}

func TestCoverageMatrixUsesCanonicalLevelsWhenLevelsOmitted(t *testing.T) {
	got, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{
		Corpus: "cefr",
		Domain: "production",
		Scales: []string{"turntaking"},
	})
	if err != nil {
		t.Fatalf("descriptors.Coverage(levels omitted) error = %v", err)
	}

	wantLevels := []string{"pre_a1", "a1", "a2", "b1", "b2", "c1", "c2"}
	if diff := cmp.Diff(wantLevels, got.Levels); diff != "" {
		t.Errorf("descriptors.Coverage(levels omitted) levels mismatch (-want +got):\n%s", diff)
	}
	if gotColumns := len(got.Columns); gotColumns != len(wantLevels) {
		t.Errorf("descriptors.Coverage(levels omitted) columns len = %d, want %d", gotColumns, len(wantLevels))
	}

	row, ok := coverageRow(got.Rows, "turntaking")
	if !ok {
		t.Fatalf("descriptors.Coverage(levels omitted) missing turntaking row in %+v", got.Rows)
	}
	if gotCells := len(row.Cells); gotCells != len(wantLevels) {
		t.Fatalf("turntaking row cells len = %d, want %d explicit cells", gotCells, len(wantLevels))
	}
	for _, emptyLevel := range []string{"pre_a1", "a2", "b2", "c1", "c2"} {
		cell, ok := coverageCell(row.Cells, emptyLevel)
		if !ok {
			t.Errorf("turntaking row missing explicit empty cell for level %q", emptyLevel)
			continue
		}
		if cell.Count != 0 || len(cell.DescriptorIDs) != 0 {
			t.Errorf("turntaking cell %q = %+v, want empty explicit cell", emptyLevel, cell)
		}
	}
	if got.GrandTotal != 2 {
		t.Errorf("descriptors.Coverage(levels omitted) grand total = %d, want 2", got.GrandTotal)
	}
}

func TestSchemaReportsUnavailableFieldWithAvailableFields(t *testing.T) {
	_, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{
		Field:   descriptors.FieldSubdomain,
		Filters: descriptors.Filters{Corpora: []string{"themes"}},
	})
	coded := requireCodedError(t, err, "unavailable_field")

	wantInvalidFilter := descriptors.InvalidFilter{Field: "field", Value: "subdomain"}
	if diff := cmp.Diff(wantInvalidFilter, coded.Details.InvalidFilter); diff != "" {
		t.Errorf("unavailable field invalidFilter mismatch (-want +got):\n%s", diff)
	}
	if fieldsContain(coded.Details.AvailableFields, descriptors.FieldSubdomain) {
		t.Errorf("availableFields = %v, want subdomain omitted for themes", coded.Details.AvailableFields)
	}
	for _, wantField := range []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldScale, descriptors.FieldLevel, descriptors.FieldCode, descriptors.FieldID} {
		if !fieldsContain(coded.Details.AvailableFields, wantField) {
			t.Errorf("availableFields = %v, want it to include %q", coded.Details.AvailableFields, wantField)
		}
	}
}

func TestSchemaSummaryValuesAndFieldErrors(t *testing.T) {
	summary, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{})
	if err != nil {
		t.Fatalf("descriptors.Schema(summary) error = %v", err)
	}
	if summary.Kind != "descriptor_schema" || len(summary.Fields) == 0 {
		t.Errorf("descriptors.Schema(summary) = %+v, want descriptor_schema with fields", summary)
	}

	values, err := descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{Field: descriptors.FieldLevel, Filters: descriptors.Filters{Corpora: []string{"cefr"}}})
	if err != nil {
		t.Fatalf("descriptors.Schema(values) error = %v", err)
	}
	wantValues := []string{"a1", "a2", "b1"}
	if values.Kind != "descriptor_schema_values" || !cmp.Equal(wantValues, values.Values) {
		t.Errorf("descriptors.Schema(values) = %+v, want kind descriptor_schema_values values %v", values, wantValues)
	}

	_, err = descriptors.Schema(context.Background(), queryDataset(), descriptors.SchemaInput{Field: descriptors.Field("subdomian")})
	coded := requireCodedError(t, err, "unknown_field")
	if diff := cmp.Diff([]string{"subdomain"}, coded.Suggestions); diff != "" {
		t.Errorf("unknown field suggestions mismatch (-want +got):\n%s", diff)
	}
	wantDetails := descriptors.ErrorDetails{
		InvalidFilter:   descriptors.InvalidFilter{Field: "field", Value: "subdomian"},
		AvailableFields: []descriptors.Field{descriptors.FieldCorpus, descriptors.FieldDomain, descriptors.FieldSubdomain, descriptors.FieldScale, descriptors.FieldLevel, descriptors.FieldCode, descriptors.FieldID},
	}
	if diff := cmp.Diff(wantDetails, coded.Details); diff != "" {
		t.Errorf("unknown field details mismatch (-want +got):\n%s", diff)
	}
}
