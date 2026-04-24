package descriptors

import (
	"context"
	"sort"
)

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
