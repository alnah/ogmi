package cli

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/alnah/ogmi/internal/specs"
)

type specsExportResult struct {
	Kind          string `json:"kind"`
	SchemaVersion string `json:"schemaVersion"`
	Output        string `json:"output"`
	Forced        bool   `json:"forced"`
}

func specsCommand(cfg *config, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "specs",
		Short:   "Manage bundled specs",
		Long:    "Manage bundled descriptor specs.",
		Example: "ogmi specs export --output ./specs",
	}
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
			result := specsExportResult{
				Kind:          "specs_export",
				SchemaVersion: SchemaVersion,
				Output:        output,
				Forced:        force,
			}
			return writeOutput(
				stdout,
				cfg.format,
				func(w io.Writer) error {
					return json.NewEncoder(w).Encode(result)
				},
				[]string{"Exported specs to " + output},
			)
		},
	}
	export.Flags().StringVar(&output, "output", "", "output directory")
	export.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	cmd.AddCommand(export)
	return cmd
}
