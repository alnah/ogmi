package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/alnah/ogmi/internal/descriptors"
	"github.com/alnah/ogmi/internal/specs"
)

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
