package xslice

import "strconv"

func IntJoin(ns []int, sep byte) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, sep)
	}
	b = b[:len(b)-1]
	return string(b)
}

func Int32Join(ns []int32, sep byte) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, sep)
	}
	b = b[:len(b)-1]
	return string(b)
}

func Int64Join(ns []int64, sep byte) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, n, 10)
		b = append(b, sep)
	}
	b = b[:len(b)-1]
	return string(b)
}
