package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type jsonWriter func(io.Writer, bool) error

func writeOutput(stdout io.Writer, format string, pretty bool, writeJSON jsonWriter, text []string) error {
	if format == "text" {
		_, err := fmt.Fprintln(stdout, strings.Join(text, "\n"))
		return err
	}
	return writeJSON(stdout, pretty)
}

func newJSONEncoder(w io.Writer, pretty bool) *json.Encoder {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder
}
