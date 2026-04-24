package descriptors

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads descriptor specs from fsys and returns a normalized dataset.
func Load(ctx context.Context, fsys fs.FS, options LoadOptions) (Dataset, error) {
	if err := ctx.Err(); err != nil {
		return Dataset{}, CodedError{Code: "internal", Message: err.Error()}
	}
	corpora, err := resolveCorpora(options.Corpora)
	if err != nil {
		return Dataset{}, err
	}

	var dataset Dataset
	for _, corpus := range corpora {
		files := descriptorFiles(fsys, corpus)
		for _, file := range files {
			part, err := loadFile(fsys, corpus, file)
			if err != nil {
				return Dataset{}, err
			}
			dataset.Scales = append(dataset.Scales, part.Scales...)
			dataset.Descriptors = append(dataset.Descriptors, part.Descriptors...)
		}
	}
	if len(dataset.Scales) == 0 && len(dataset.Descriptors) == 0 {
		return Dataset{}, CodedError{Code: "missing_specs", Message: "Descriptor specs not found"}
	}
	sort.SliceStable(dataset.Scales, func(i, j int) bool { return compareScale(dataset.Scales[i], dataset.Scales[j]) < 0 })
	sort.SliceStable(dataset.Descriptors, func(i, j int) bool { return compareDescriptor(dataset.Descriptors[i], dataset.Descriptors[j]) < 0 })
	return dataset, nil
}

type yamlFile struct {
	Catalog []yamlRow `yaml:"catalog"`
	Entries []yamlRow `yaml:"entries"`
}

type yamlRow struct {
	Scale       string      `yaml:"scale"`
	Level       string      `yaml:"level"`
	Code        string      `yaml:"code"`
	ID          string      `yaml:"id"`
	Description description `yaml:"description"`
}

type description []string

func (d *description) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var text string
		if err := value.Decode(&text); err != nil {
			return err
		}
		*d = []string{text}
		return nil
	case yaml.SequenceNode:
		var lines []string
		if err := value.Decode(&lines); err != nil {
			return err
		}
		*d = lines
		return nil
	default:
		return fmt.Errorf("description must be string or list")
	}
}

func descriptorFiles(fsys fs.FS, corpus Corpus) []string {
	seen := make(map[string]bool)
	files := []string{}
	for _, root := range corpus.Roots {
		if _, err := fs.Stat(fsys, root); err != nil {
			continue
		}
		_ = fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil {
				return fs.SkipDir
			}
			if !entry.IsDir() && path.Base(name) == "descriptors.yml" && !seen[name] {
				seen[name] = true
				files = append(files, name)
			}
			return nil
		})
	}
	for _, file := range corpus.Files {
		if _, err := fs.Stat(fsys, file); err == nil && !seen[file] {
			seen[file] = true
			files = append(files, file)
		}
	}
	sort.Strings(files)
	return files
}

func loadFile(fsys fs.FS, corpus Corpus, file string) (Dataset, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return Dataset{}, CodedError{Code: "missing_specs", Message: fmt.Sprintf("Descriptor spec %s not found", file)}
	}
	var raw yamlFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Dataset{}, CodedError{Code: "invalid_yaml", Message: fmt.Sprintf("Invalid descriptor YAML in %s: %v", file, err)}
	}
	values := pathValues(corpus, file)
	var dataset Dataset
	for index, row := range raw.Catalog {
		code := normalizeToken(row.Code)
		id := strings.TrimSpace(row.ID)
		desc := normalizeDescription([]string(row.Description))
		if code == "" || id == "" || len(desc) == 0 {
			return Dataset{}, CodedError{Code: "invalid_row", Message: fmt.Sprintf("Invalid descriptor catalog row %d in %s", index+1, file)}
		}
		dataset.Scales = append(dataset.Scales, DescriptorScaleRecord{Corpus: corpus.Name, Domain: values.Domain, Subdomain: values.Subdomain, Code: code, ID: id, Description: desc, File: file})
	}
	for index, row := range raw.Entries {
		scale := normalizeToken(row.Scale)
		level := normalizeLevel(row.Level)
		code := normalizeToken(row.Code)
		id := strings.TrimSpace(row.ID)
		desc := normalizeDescription([]string(row.Description))
		if scale == "" || level == "" || code == "" || id == "" || len(desc) == 0 {
			return Dataset{}, CodedError{Code: "invalid_row", Message: fmt.Sprintf("Invalid descriptor entries row %d in %s", index+1, file)}
		}
		dataset.Descriptors = append(dataset.Descriptors, DescriptorRecord{Corpus: corpus.Name, Domain: values.Domain, Subdomain: values.Subdomain, Scale: scale, Level: level, Code: code, ID: id, Description: strings.Join(desc, "\n"), File: file})
	}
	return dataset, nil
}

type pathFieldValues struct{ Domain, Subdomain string }

func pathValues(corpus Corpus, file string) pathFieldValues {
	for _, root := range corpus.Roots {
		rel, ok := strings.CutPrefix(file, strings.TrimSuffix(root, "/")+"/")
		if !ok {
			continue
		}
		segments := strings.Split(path.Dir(rel), "/")
		if path.Dir(rel) == "." {
			segments = nil
		}
		values := pathFieldValues{}
		if len(corpus.PathFields) > 0 && len(segments) > 0 && corpus.PathFields[0] == FieldDomain {
			values.Domain = normalizeToken(segments[0])
		}
		if len(corpus.PathFields) > 1 && len(segments) > 1 && corpus.PathFields[1] == FieldSubdomain {
			values.Subdomain = normalizeToken(segments[1])
		}
		return values
	}
	return pathFieldValues{}
}
