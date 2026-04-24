package descriptors

import "strings"

var canonicalLevels = []string{"pre_a1", "a1", "a2", "b1", "b2", "c1", "c2"}

func compareLevels(left, right string) int {
	li, lok := levelIndex(left)
	ri, rok := levelIndex(right)
	if lok && rok {
		return li - ri
	}
	if lok {
		return -1
	}
	if rok {
		return 1
	}
	return strings.Compare(left, right)
}

func levelIndex(level string) (int, bool) {
	for index, candidate := range canonicalLevels {
		if candidate == level {
			return index, true
		}
	}
	return 0, false
}

func isCanonicalLevel(level string) bool { _, ok := levelIndex(level); return ok }
