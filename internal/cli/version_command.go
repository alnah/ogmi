package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

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
