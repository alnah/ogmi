package cli

import (
	"errors"
	"strings"

	"github.com/alnah/ogmi/internal/descriptors"
)

type usageError struct{ message string }

func (e usageError) Error() string { return e.message }

type errorEnvelope struct {
	Kind          string                 `json:"kind"`
	SchemaVersion string                 `json:"schemaVersion"`
	Error         descriptors.CodedError `json:"error"`
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
