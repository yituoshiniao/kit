package xstring

import (
	"strconv"
	"strings"
)

// SplitToInt64 切割字符串为int64
// 例如 111_111 为 [111,111]
func SplitToInt64(s, sep string) ([]int64, error) {
	ss := strings.Split(s, sep)
	r := make([]int64, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			return nil, err
		}

		r[i] = v
	}
	return r, nil
}
