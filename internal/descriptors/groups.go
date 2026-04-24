package descriptors

import "strings"

func buildGroups(items []DescriptorRecord, fields []Field) []Group {
	if len(fields) == 0 {
		return nil
	}
	groups := []Group{}
	indexes := make(map[string]int)
	for _, item := range items {
		key := make(map[Field]string)
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			value := recordField(item, field)
			if value != "" {
				key[field] = value
			}
			parts = append(parts, string(field)+"="+value)
		}
		id := strings.Join(parts, "\x00")
		index, ok := indexes[id]
		if !ok {
			indexes[id] = len(groups)
			groups = append(groups, Group{KeyFields: fields, Key: key, Total: 1, Items: []DescriptorRecord{item}})
			continue
		}
		groups[index].Total++
		groups[index].Items = append(groups[index].Items, item)
	}
	return groups
}
