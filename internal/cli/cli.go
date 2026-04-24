package cli

import (
	"context"
	"fmt"
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
	_ = ctx
	_ = args
	_ = stdout
	_, _ = fmt.Fprintln(stderr, `{"kind":"error","schemaVersion":"v1","error":{"code":"internal","message":"not implemented"}}`)
	return ExitInternal
}
