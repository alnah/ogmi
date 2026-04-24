package cli

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type config struct {
	format string
	specs  string
}

func newRootCommand(ctx context.Context, cfg *config, stdout io.Writer) *cobra.Command {
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
