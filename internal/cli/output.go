package cli

import (
	"fmt"
	"io"
	"strings"
)

type jsonWriter func(io.Writer) error

func writeOutput(stdout io.Writer, format string, writeJSON jsonWriter, text []string) error {
	if format == "text" {
		_, err := fmt.Fprintln(stdout, strings.Join(text, "\n"))
		return err
	}
	return writeJSON(stdout)
}
