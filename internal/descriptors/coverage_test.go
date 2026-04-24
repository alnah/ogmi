package descriptors_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alnah/ogmi/internal/descriptors"
)

func TestCoverageMatrixReportsRichRowsColumnsCellsAndContinuity(t *testing.T) {
	got, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{
		Corpus: "cefr",
		Domain: "production",
		Scales: []string{"addressing_audiences", "turntaking"},
		Levels: []string{"a1", "a2", "b1"},
	})
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

	wantColumns := []descriptors.CoverageColumn{
		{Level: "a1", Total: 2},
		{Level: "a2", Total: 1},
		{Level: "b1", Total: 1},
	}
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
				{
					Scale:         "addressing_audiences",
					Level:         "a1",
					Count:         1,
					DescriptorIDs: []string{deliverToastDescriptorID},
				},
				{
					Scale:         "addressing_audiences",
					Level:         "a2",
					Count:         1,
					DescriptorIDs: []string{presentSimpleAnnouncementDescriptorID},
				},
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
				{Scale: "turntaking", Level: "a1", Count: 1, DescriptorIDs: []string{openSimpleExchangeDescriptorID}},
				{Scale: "turntaking", Level: "a2", Count: 0, DescriptorIDs: nil},
				{Scale: "turntaking", Level: "b1", Count: 1, DescriptorIDs: []string{maintainExchangeDescriptorID}},
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

func TestCoverageMatrixDerivesSortedScaleRowsWhenScalesOmitted(t *testing.T) {
	got, err := descriptors.Coverage(context.Background(), queryDataset(), descriptors.CoverageInput{
		Corpus: "cefr",
		Domain: "production",
		Levels: []string{"a1"},
	})
	if err != nil {
		t.Fatalf("descriptors.Coverage(scales omitted) error = %v", err)
	}
	wantScales := []string{"addressing_audiences", "turntaking"}
	if diff := cmp.Diff(wantScales, got.Scales); diff != "" {
		t.Errorf("descriptors.Coverage(scales omitted) scales mismatch (-want +got):\n%s", diff)
	}
	if got.GrandTotal != 2 {
		t.Errorf("descriptors.Coverage(scales omitted) grand total = %d, want 2", got.GrandTotal)
	}
}
