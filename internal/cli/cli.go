package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alnah/ogmi/internal/descriptors"
	"github.com/alnah/ogmi/internal/specs"
)

const (
	ExitSuccess    = 0
	ExitDomain     = 1
	ExitUsage      = 2
	ExitInternal   = 3
	SchemaVersion  = "v1"
	DefaultVersion = "dev"
)

type config struct {
	format string
	specs  string
}

type usageError struct{ message string }

func (e usageError) Error() string { return e.message }

type errorEnvelope struct {
	Kind          string                 `json:"kind"`
	SchemaVersion string                 `json:"schemaVersion"`
	Error         descriptors.CodedError `json:"error"`
}

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

type specsExportResult struct {
	Kind          string `json:"kind"`
	SchemaVersion string `json:"schemaVersion"`
	Output        string `json:"output"`
	Forced        bool   `json:"forced"`
}

type examplesResult struct {
	Kind          string   `json:"kind"`
	SchemaVersion string   `json:"schemaVersion"`
	Examples      []string `json:"examples"`
}

// Run executes ogmi and returns the process exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg := &config{format: "json"}
	root := newRootCommand(ctx, cfg, stdout, stderr)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SilenceUsage = true
	root.SilenceErrors = true

	if err := root.ExecuteContext(ctx); err != nil {
		coded, exitCode := classifyError(err)
		_ = json.NewEncoder(stderr).Encode(errorEnvelope{Kind: "error", SchemaVersion: SchemaVersion, Error: coded})
		return exitCode
	}
	return ExitSuccess
}

func newRootCommand(ctx context.Context, cfg *config, stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   "ogmi",
		Short: "Ogmi language descriptor CLI",
		Long:  "Ogmi queries bundled language descriptor specs.\n\nCommon Commands:\n  ogmi descriptors list\n  ogmi descriptors corpora\n  ogmi specs export\n  ogmi version",
		Example: strings.Join([]string{
			"ogmi descriptors corpora",
			"ogmi descriptors list --corpus cefr --level a1 --limit 5",
			"ogmi specs export --output ./specs",
		}, "\n"),
	}
	root.PersistentFlags().StringVar(&cfg.format, "format", "json", "output format: json or text")
	root.PersistentFlags().StringVar(&cfg.specs, "specs", "", "descriptor specs root; overrides OGMI_SPECS and embedded specs")
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_ = cmd
		return usageError{message: err.Error()}
	})
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		_ = args
		if cmd.Name() == "help" {
			return nil
		}
		cfg.format = strings.ToLower(strings.TrimSpace(cfg.format))
		if cfg.format != "json" && cfg.format != "text" {
			return usageError{message: "--format must be json or text"}
		}
		return nil
	}

	root.AddCommand(versionCommand(stdout))
	root.AddCommand(specsCommand(cfg, stdout))
	root.AddCommand(descriptorsCommand(ctx, cfg, stdout))
	return root
}

func versionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the Ogmi version",
		Long:    "Print the Ogmi CLI version and exit.",
		Example: "ogmi version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			_, err := fmt.Fprintf(stdout, "ogmi version %s\n", DefaultVersion)
			return err
		},
	}
}

func specsCommand(cfg *config, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "specs", Short: "Manage bundled specs", Long: "Manage bundled descriptor specs.", Example: "ogmi specs export --output ./specs"}
	var output string
	var force bool
	export := &cobra.Command{
		Use:     "export",
		Short:   "Export bundled specs",
		Long:    "Export embedded descriptor specs to a directory, preserving .gitkeep files.",
		Example: "ogmi specs export --output ./specs --force",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			if output == "" {
				return usageError{message: "specs export requires --output"}
			}
			if err := specs.Export(specs.Bundle(), output, force); err != nil {
				return err
			}
			result := specsExportResult{Kind: "specs_export", SchemaVersion: SchemaVersion, Output: output, Forced: force}
			return writeOutput(stdout, cfg.format, func(w io.Writer) error { return json.NewEncoder(w).Encode(result) }, []string{"Exported specs to " + output})
		},
	}
	export.Flags().StringVar(&output, "output", "", "output directory")
	export.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	cmd.AddCommand(export)
	return cmd
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
			if filters.SortOrder != "" && strings.ToLower(filters.SortOrder) != "asc" && strings.ToLower(filters.SortOrder) != "desc" {
				return usageError{message: "--sort-order must be asc or desc"}
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

func loadDataset(ctx context.Context, cfg *config, corpora []string) (descriptors.Dataset, error) {
	source, err := specs.Resolve(specs.SourceRequest{FlagPath: cfg.specs, EnvPath: os.Getenv("OGMI_SPECS")})
	if err != nil {
		return descriptors.Dataset{}, err
	}
	dataset, err := descriptors.Load(ctx, source.FS, descriptors.LoadOptions{Corpora: corpora})
	if err == nil {
		return dataset, nil
	}
	var coded descriptors.CodedError
	if errors.As(err, &coded) && coded.Code == "missing_specs" && source.Description != "embedded" {
		coded.Message = fmt.Sprintf("Descriptor specs not found: %s", source.Description)
		return descriptors.Dataset{}, coded
	}
	return descriptors.Dataset{}, err
}

type jsonWriter func(io.Writer) error

func writeOutput(stdout io.Writer, format string, writeJSON jsonWriter, text []string) error {
	if format == "text" {
		_, err := fmt.Fprintln(stdout, strings.Join(text, "\n"))
		return err
	}
	return writeJSON(stdout)
}

func classifyError(err error) (descriptors.CodedError, int) {
	var coded descriptors.CodedError
	if errors.As(err, &coded) {
		return coded, ExitDomain
	}
	var usage usageError
	if errors.As(err, &usage) {
		return descriptors.CodedError{Code: "usage", Message: usage.Error()}, ExitUsage
	}
	message := err.Error()
	if strings.Contains(message, "unknown command") || strings.Contains(message, "requires") {
		return descriptors.CodedError{Code: "usage", Message: message}, ExitUsage
	}
	return descriptors.CodedError{Code: "internal", Message: message}, ExitInternal
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
