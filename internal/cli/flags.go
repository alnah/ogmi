package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/alnah/ogmi/internal/descriptors"
)

func bindListFlags(cmd *cobra.Command, filters *descriptors.Filters) {
	cmd.Flags().StringVar(&filters.Query, "query", "", "text query")
	cmd.Flags().StringArrayVar(&filters.Corpora, "corpus", nil, "corpus filter")
	cmd.Flags().StringVar(&filters.Domain, "domain", "", "domain filter")
	cmd.Flags().StringVar(&filters.Subdomain, "subdomain", "", "subdomain filter")
	cmd.Flags().StringArrayVar(&filters.Scales, "scale", nil, "scale filter")
	cmd.Flags().StringArrayVar(&filters.Levels, "level", nil, "level filter")
	cmd.Flags().StringVar(&filters.Code, "code", "", "code filter")
	cmd.Flags().StringVar(&filters.ID, "id", "", "id filter")
	cmd.Flags().IntVar(&filters.Offset, "offset", 0, "pagination offset")
	cmd.Flags().IntVar(&filters.Limit, "limit", 0, "pagination limit")
	cmd.Flags().Var((*fieldValue)(&filters.SortBy), "sort-by", "sort field")
	cmd.Flags().StringVar(&filters.SortOrder, "sort-order", "asc", "sort order: asc or desc")
	cmd.Flags().Var((*fieldSliceValue)(&filters.GroupBy), "group-by", "group field")
	cmd.Flags().BoolVar(&filters.WithFacets, "facets", false, "include facets")
}

func bindScaleFlags(cmd *cobra.Command, filters *descriptors.Filters) {
	cmd.Flags().StringArrayVar(&filters.Corpora, "corpus", nil, "corpus filter")
	cmd.Flags().StringVar(&filters.Domain, "domain", "", "domain filter")
	cmd.Flags().StringVar(&filters.Subdomain, "subdomain", "", "subdomain filter")
	cmd.Flags().StringArrayVar(&filters.Scales, "scale", nil, "scale filter")
	cmd.Flags().StringVar(&filters.ID, "id", "", "id filter")
	cmd.Flags().StringVar(&filters.Query, "query", "", "text query")
	cmd.Flags().IntVar(&filters.Offset, "offset", 0, "pagination offset")
	cmd.Flags().IntVar(&filters.Limit, "limit", 0, "pagination limit")
	cmd.Flags().Var((*fieldValue)(&filters.SortBy), "sort-by", "sort field")
	cmd.Flags().StringVar(&filters.SortOrder, "sort-order", "asc", "sort order")
}

type fieldValue descriptors.Field

func (v *fieldValue) String() string { return string(*v) }
func (v *fieldValue) Set(value string) error {
	*v = fieldValue(descriptors.Field(strings.TrimSpace(value)))
	return nil
}
func (v *fieldValue) Type() string { return "field" }

type fieldSliceValue []descriptors.Field

func (v *fieldSliceValue) String() string {
	fields := make([]string, 0, len(*v))
	for _, field := range *v {
		fields = append(fields, string(field))
	}
	return strings.Join(fields, ",")
}
func (v *fieldSliceValue) Set(value string) error {
	*v = append(*v, descriptors.Field(strings.TrimSpace(value)))
	return nil
}
func (v *fieldSliceValue) Type() string { return "field" }
