package xslice

import "strconv"

// Int64ToString int64分片 转为 string分片
func Int64ToString(in []int64) (out []string) {
	for _, i := range in {
		out = append(out, strconv.FormatInt(i, 10))
	}

	return
}
