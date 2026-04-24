package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alnah/ogmi/internal/descriptors"
)

type corporaResult struct {
	Kind          string               `json:"kind"`
	SchemaVersion string               `json:"schemaVersion"`
	Corpora       []descriptors.Corpus `json:"corpora"`
}

type fieldsResult struct {
	Kind          string             `json:"kind"`
	SchemaVersion string             `json:"schemaVersion"`
	Fields        []fieldDescription `json:"fields"`
}

type fieldDescription struct {
	Name       descriptors.Field `json:"name"`
	Type       string            `json:"type"`
	Filterable bool              `json:"filterable"`
	Groupable  bool              `json:"groupable"`
}

type examplesResult struct {
	Kind          string   `json:"kind"`
	SchemaVersion string   `json:"schemaVersion"`
	Examples      []string `json:"examples"`
}

func descriptorsCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descriptors",
		Short: "Query language descriptors",
		Long:  "Query language descriptors. Workflow: inspect corpora, inspect fields or schema, list descriptors, then get descriptor by id.",
		Example: strings.Join([]string{
			"ogmi descriptors corpora",
			"ogmi descriptors fields --corpus cefr",
			"ogmi descriptors schema --field level --corpus cefr",
			"ogmi descriptors list --corpus cefr --domain production --subdomain speaking --level a1",
			"ogmi descriptors get --corpus cefr --id cefr.production.speaking.descriptors.addressing_audiences.use_very_short_prepared_text_to_deliver_rehearsed_statement.a1",
		}, "\n"),
		Args: rejectUnknownDescriptorCommand,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			return cmd.Help()
		},
	}
	cmd.AddCommand(corporaCommand(ctx, cfg, stdout))
	cmd.AddCommand(fieldsCommand(cfg, stdout))
	cmd.AddCommand(schemaCommand(ctx, cfg, stdout))
	cmd.AddCommand(listCommand(ctx, cfg, stdout))
	cmd.AddCommand(scalesCommand(ctx, cfg, stdout))
	cmd.AddCommand(getCommand(ctx, cfg, stdout))
	cmd.AddCommand(compareCommand(ctx, cfg, stdout))
	cmd.AddCommand(coverageCommand(ctx, cfg, stdout))
	cmd.AddCommand(examplesCommand(cfg, stdout))
	return cmd
}

func corporaCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "corpora",
		Short:   "List descriptor corpora",
		Long:    "List registered descriptor corpora and their source metadata.",
		Example: "ogmi descriptors corpora",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			if _, err := loadDataset(ctx, cfg, nil); err != nil {
				return err
			}
			result := corporaResult{Kind: "descriptor_corpora", SchemaVersion: SchemaVersion, Corpora: descriptors.Registry()}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, corporaText(result))
		},
	}
}

func fieldsCommand(cfg *config, stdout io.Writer) *cobra.Command {
	var corpus string
	cmd := &cobra.Command{
		Use:     "fields",
		Short:   "List descriptor fields",
		Long:    "List known descriptor fields and their query roles.",
		Example: "ogmi descriptors fields --corpus cefr",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			_ = corpus
			fields := []fieldDescription{
				{Name: descriptors.FieldCorpus, Type: "token", Filterable: true, Groupable: true},
				{Name: descriptors.FieldDomain, Type: "token", Filterable: true, Groupable: true},
				{Name: descriptors.FieldSubdomain, Type: "token", Filterable: true, Groupable: true},
				{Name: descriptors.FieldScale, Type: "token", Filterable: true, Groupable: true},
				{Name: descriptors.FieldLevel, Type: "level", Filterable: true, Groupable: true},
				{Name: descriptors.FieldCode, Type: "token", Filterable: true, Groupable: false},
				{Name: descriptors.FieldID, Type: "id", Filterable: true, Groupable: false},
			}
			result := fieldsResult{Kind: "descriptor_fields", SchemaVersion: SchemaVersion, Fields: fields}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{"Descriptor fields: corpus, domain, subdomain, scale, level, code, id"})
		},
	}
	cmd.Flags().StringVar(&corpus, "corpus", "", "corpus filter")
	return cmd
}

func schemaCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var input descriptors.SchemaInput
	var field string
	bindFilterFlags := func(cmd *cobra.Command, filters *descriptors.Filters) {
		cmd.Flags().StringArrayVar(&filters.Corpora, "corpus", nil, "corpus filter")
		cmd.Flags().StringVar(&filters.Domain, "domain", "", "domain filter")
		cmd.Flags().StringVar(&filters.Subdomain, "subdomain", "", "subdomain filter")
		cmd.Flags().StringArrayVar(&filters.Scales, "scale", nil, "scale filter")
		cmd.Flags().StringArrayVar(&filters.Levels, "level", nil, "level filter")
		cmd.Flags().StringVar(&filters.Code, "code", "", "code filter")
		cmd.Flags().StringVar(&filters.ID, "id", "", "id filter")
	}
	cmd := &cobra.Command{
		Use:     "schema",
		Short:   "Describe descriptor schema",
		Long:    "Describe descriptor schema or list values for one field.",
		Example: "ogmi descriptors schema --field level --corpus cefr",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			input.Field = descriptors.Field(field)
			dataset, err := loadDataset(ctx, cfg, input.Corpora)
			if err != nil {
				return err
			}
			result, err := descriptors.Schema(ctx, dataset, input)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{"Descriptor schema"})
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "schema field")
	bindFilterFlags(cmd, &input.Filters)
	return cmd
}

func listCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var filters descriptors.Filters
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List descriptors",
		Long:    "List descriptors with filters, sorting, pagination, groups, and facets.",
		Example: "ogmi descriptors list --corpus cefr --domain production --subdomain speaking --level a1",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			if err := validateSortOrder(filters.SortOrder); err != nil {
				return err
			}
			if err := validateListSortBy(filters.SortBy); err != nil {
				return err
			}
			dataset, err := loadDataset(ctx, cfg, filters.Corpora)
			if err != nil {
				return err
			}
			result, err := descriptors.Query(ctx, dataset, filters)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{fmt.Sprintf("Descriptors: %d", result.Total)})
		},
	}
	bindListFlags(cmd, &filters)
	return cmd
}

func scalesCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var filters descriptors.Filters
	cmd := &cobra.Command{
		Use:     "scales",
		Short:   "List descriptor scales",
		Long:    "List descriptor scale records for one or more corpora.",
		Example: "ogmi descriptors scales --corpus cefr --query audience",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			if err := validateSortOrder(filters.SortOrder); err != nil {
				return err
			}
			if err := validateScaleSortBy(filters.SortBy); err != nil {
				return err
			}
			dataset, err := loadDataset(ctx, cfg, filters.Corpora)
			if err != nil {
				return err
			}
			result, err := descriptors.QueryScales(ctx, dataset, filters)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{fmt.Sprintf("Descriptor scales: %d", result.Total)})
		},
	}
	bindScaleFlags(cmd, &filters)
	return cmd
}

func getCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var input descriptors.GetInput
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get one descriptor",
		Long:    "Get one descriptor by id or by a unique code within a corpus.",
		Example: "ogmi descriptors get --corpus cefr --id cefr.production.speaking.descriptors.addressing_audiences.use_very_short_prepared_text_to_deliver_rehearsed_statement.a1",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			if strings.TrimSpace(input.ID) == "" && strings.TrimSpace(input.Code) == "" {
				return usageError{message: "descriptors get requires --id or --code"}
			}
			dataset, err := loadDataset(ctx, cfg, []string{input.Corpus})
			if err != nil {
				return err
			}
			result, err := descriptors.Get(ctx, dataset, input)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{"Descriptor: " + result.Descriptor.ID})
		},
	}
	cmd.Flags().StringVar(&input.Corpus, "corpus", "", "corpus filter")
	cmd.Flags().StringVar(&input.ID, "id", "", "descriptor id")
	cmd.Flags().StringVar(&input.Code, "code", "", "descriptor code")
	cmd.Flags().StringVar(&input.Domain, "domain", "", "domain filter")
	cmd.Flags().StringVar(&input.Subdomain, "subdomain", "", "subdomain filter")
	cmd.Flags().StringVar(&input.Scale, "scale", "", "scale filter")
	cmd.Flags().StringVar(&input.Level, "level", "", "level filter")
	return cmd
}

func compareCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var input descriptors.CompareLevelsInput
	cmd := &cobra.Command{
		Use:     "compare-levels",
		Short:   "Compare descriptor levels",
		Long:    "Compare descriptors for one scale across CEFR levels.",
		Example: "ogmi descriptors compare-levels --corpus cefr --scale turntaking --level a1 --level b1",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			dataset, err := loadDataset(ctx, cfg, []string{input.Corpus})
			if err != nil {
				return err
			}
			result, err := descriptors.CompareLevels(ctx, dataset, input)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{fmt.Sprintf("Compared levels: %d", len(result.Levels))})
		},
	}
	cmd.Flags().StringVar(&input.Corpus, "corpus", "", "corpus filter")
	cmd.Flags().StringVar(&input.Scale, "scale", "", "scale filter")
	cmd.Flags().StringVar(&input.Domain, "domain", "", "domain filter")
	cmd.Flags().StringVar(&input.Subdomain, "subdomain", "", "subdomain filter")
	cmd.Flags().StringArrayVar(&input.Levels, "level", nil, "level filter")
	cmd.Flags().IntVar(&input.Limit, "limit", 0, "item limit")
	cmd.Flags().StringVar(&input.Query, "query", "", "text query")
	return cmd
}

func coverageCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
	var input descriptors.CoverageInput
	cmd := &cobra.Command{
		Use:     "coverage",
		Short:   "Build a coverage matrix",
		Long:    "Build a descriptor coverage matrix by scale and level.",
		Example: "ogmi descriptors coverage --corpus cefr --domain production",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			dataset, err := loadDataset(ctx, cfg, []string{input.Corpus})
			if err != nil {
				return err
			}
			result, err := descriptors.Coverage(ctx, dataset, input)
			if err != nil {
				return err
			}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{fmt.Sprintf("Coverage total: %d", result.GrandTotal)})
		},
	}
	cmd.Flags().StringVar(&input.Corpus, "corpus", "", "corpus filter")
	cmd.Flags().StringVar(&input.Domain, "domain", "", "domain filter")
	cmd.Flags().StringVar(&input.Subdomain, "subdomain", "", "subdomain filter")
	cmd.Flags().StringArrayVar(&input.Scales, "scale", nil, "scale filter")
	cmd.Flags().StringArrayVar(&input.Levels, "level", nil, "level filter")
	return cmd
}

func examplesCommand(cfg *config, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "examples",
		Short:   "Show descriptor workflow examples",
		Long:    "Show practical descriptor command examples for agents and humans.",
		Example: "ogmi descriptors examples",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			result := examplesResult{Kind: "descriptor_examples", SchemaVersion: SchemaVersion, Examples: descriptorExamples()}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, result.Examples)
		},
	}
}

func rejectUnknownDescriptorCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return usageError{message: fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())}
}

func validateSortOrder(value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "asc", "desc":
		return nil
	default:
		return usageError{message: "--sort-order must be asc or desc"}
	}
}

func validateListSortBy(value descriptors.Field) error {
	return validateSortBy(value, map[descriptors.Field]bool{
		descriptors.FieldCorpus:    true,
		descriptors.FieldDomain:    true,
		descriptors.FieldSubdomain: true,
		descriptors.FieldScale:     true,
		descriptors.FieldLevel:     true,
		descriptors.FieldCode:      true,
		descriptors.FieldID:        true,
	})
}

func validateScaleSortBy(value descriptors.Field) error {
	return validateSortBy(value, map[descriptors.Field]bool{
		descriptors.FieldCorpus:    true,
		descriptors.FieldDomain:    true,
		descriptors.FieldSubdomain: true,
		descriptors.FieldScale:     true,
		descriptors.FieldCode:      true,
		descriptors.FieldID:        true,
	})
}

func validateSortBy(value descriptors.Field, allowed map[descriptors.Field]bool) error {
	field := descriptors.Field(strings.ToLower(strings.TrimSpace(string(value))))
	if field == "" || allowed[field] {
		return nil
	}
	return usageError{message: fmt.Sprintf("--sort-by %s is not supported", field)}
}

func corporaText(result corporaResult) []string {
	lines := make([]string, 0, len(result.Corpora))
	for _, corpus := range result.Corpora {
		lines = append(lines, corpus.Name)
	}
	return lines
}

func descriptorExamples() []string {
	return []string{
		"ogmi descriptors corpora",
		"ogmi descriptors list --corpus cefr --domain production --subdomain speaking --level a1",
		"ogmi descriptors get --corpus cefr --id cefr.production.speaking.descriptors.addressing_audiences.use_very_short_prepared_text_to_deliver_rehearsed_statement.a1",
		"ogmi descriptors scales --corpus cefr",
		"ogmi descriptors schema --field level --corpus cefr",
		"ogmi descriptors coverage --corpus cefr",
		"ogmi specs export --output ./specs && ogmi --specs ./specs descriptors list --corpus themes",
	}
}
