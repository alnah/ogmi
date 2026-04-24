package cli

import (
	"context"
	"encoding/json"
	"io"
)

const (
	ExitSuccess    = 0
	ExitDomain     = 1
	ExitUsage      = 2
	ExitInternal   = 3
	SchemaVersion  = "v1"
	DefaultVersion = "dev"
)

// Run executes ogmi and returns the process exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg := &config{format: "json"}
	root := newRootCommand(ctx, cfg, stdout)
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
