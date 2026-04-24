package descriptors

import "strings"

func compareDescriptorForQuery(left, right DescriptorRecord, field Field, order string) int {
	if field == "" {
		return compareDescriptor(left, right)
	}
	base := compareRecordField(left, right, field)
	if base != 0 && order == "desc" {
		return -base
	}
	if base != 0 {
		return base
	}
	return compareDescriptor(left, right)
}

func compareScaleForQuery(left, right DescriptorScaleRecord, field Field, order string) int {
	if field == "" {
		return compareScale(left, right)
	}
	base := strings.Compare(scaleField(left, field), scaleField(right, field))
	if base != 0 && order == "desc" {
		return -base
	}
	if base != 0 {
		return base
	}
	return compareScale(left, right)
}

func compareDescriptor(left, right DescriptorRecord) int {
	comparisons := []int{
		strings.Compare(left.Corpus, right.Corpus),
		strings.Compare(left.Domain, right.Domain),
		strings.Compare(left.Subdomain, right.Subdomain),
		strings.Compare(left.Scale, right.Scale),
		compareLevels(left.Level, right.Level),
		strings.Compare(left.Code, right.Code),
		strings.Compare(left.ID, right.ID),
	}
	for _, diff := range comparisons {
		if diff != 0 {
			return diff
		}
	}
	return 0
}

func compareScale(left, right DescriptorScaleRecord) int {
	comparisons := []int{
		strings.Compare(left.Corpus, right.Corpus),
		strings.Compare(left.Domain, right.Domain),
		strings.Compare(left.Subdomain, right.Subdomain),
		strings.Compare(left.Code, right.Code),
		strings.Compare(left.ID, right.ID),
	}
	for _, diff := range comparisons {
		if diff != 0 {
			return diff
		}
	}
	return 0
}

func compareRecordField(left, right DescriptorRecord, field Field) int {
	if field == FieldLevel {
		return compareLevels(left.Level, right.Level)
	}
	return strings.Compare(recordField(left, field), recordField(right, field))
}

func recordField(record DescriptorRecord, field Field) string {
	switch field {
	case FieldCorpus:
		return record.Corpus
	case FieldDomain:
		return record.Domain
	case FieldSubdomain:
		return record.Subdomain
	case FieldScale:
		return record.Scale
	case FieldLevel:
		return record.Level
	case FieldCode:
		return record.Code
	case FieldID:
		return record.ID
	default:
		return ""
	}
}

func scaleField(record DescriptorScaleRecord, field Field) string {
	switch field {
	case FieldCorpus:
		return record.Corpus
	case FieldDomain:
		return record.Domain
	case FieldSubdomain:
		return record.Subdomain
	case FieldScale, FieldCode:
		return record.Code
	case FieldID:
		return record.ID
	default:
		return ""
	}
}
