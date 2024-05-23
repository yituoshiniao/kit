package v1

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/yituoshiniao/kit/xslice"
)

func ScanJoinStrtoInt64(val interface{}) ([]int64, error) {
	strval, ok := val.(string)
	if !ok || val == "" {
		return nil, nil
	}

	ss := strings.Split(strval, ",")

	var ret []int64
	for _, s := range ss {
		if s == "" {
			continue
		}

		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		ret = append(ret, i)
	}
	return ret, nil
}

// CommaInt64Slice 模型类型
// 当存储时 []int64{1,2} => 1,2
// 到读取时 1,2 => []int64{1,2}
type CommaInt64Slice []int64

func (j CommaInt64Slice) Value() (driver.Value, error) {
	return xslice.Int64Join(j, ','), nil
}
func (j *CommaInt64Slice) Scan(val interface{}) (err error) {
	ss, err := ScanJoinStrtoInt64(val)
	if err != nil {
		return err
	}
	*j = *&ss
	return nil
}
