package xslice

// Int64Contains tells whether haystack contains needle.
func Int64Contains(haystack []int64, needle int64) bool {
	for _, n := range haystack {
		if needle == n {
			return true
		}
	}
	return false
}

// StringContains tells whether haystack contains needle.
func StringContains(haystack []string, needle string) bool {
	for _, n := range haystack {
		if needle == n {
			return true
		}
	}
	return false
}
