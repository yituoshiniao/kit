package xslice

func UniqueInt64(in []int64) (out []int64) {
	if len(in) == 0 {
		return
	}

	keys := make(map[int64]struct{})
	for _, entry := range in {
		if _, ok := keys[entry]; !ok {
			keys[entry] = struct{}{}
			out = append(out, entry)
		}
	}

	return out
}
