package descriptors

func stringIn(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func fieldIn(value Field, values []Field) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
