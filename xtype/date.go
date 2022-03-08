//go:generate msgp -tests=false
package xtype

import (
	"gitlab.intsig.net/cs-server2/kit-test-cb/xtime"
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	ScanError = "Scan: 不能转换 %T 为 *DataTime 类型."

	zeroTimeValue = "0000-00-00 00:00:00"
)

var ErrDateTimeIsEmpty = errors.New("传入的时间是空串，无法转为*gtype.DateTime类型")

type DateTime struct {
	time.Time
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == `"0000-00-00 00:00:00"` {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.ParseInLocation(xtime.Fmt, s, time.Local)
	if err != nil {
		// find last "."
		i := strings.LastIndex(s, ".")
		if i == -1 {
			return err
		}

		// try to parse the slice of the string
		t, err = time.ParseInLocation(xtime.Fmt, s[0:i], time.Local)
		if err != nil {
			return err
		}
	}
	d.Time = t
	return nil
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *DateTime) Scan(value interface{}) error {
	switch src := value.(type) {
	case nil:
		return nil
	case string:
		if src == zeroTimeValue {
			return nil
		}
		t, err := time.ParseInLocation(xtime.Fmt, src, time.Local)
		if err != nil {
			// find last "."
			i := strings.LastIndex(src, ".")
			if i == -1 {
				return errors.WithStack(err)
			}

			// try to parse the slice of the string
			t, err = time.ParseInLocation(xtime.Fmt, src[0:i], time.Local)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		d.Time = t
		return nil
	case time.Time:
		if !src.IsZero() {
			d.Time = src
		}
		return nil
	default:
		return errors.Errorf(ScanError, src)
	}
}

func (d *DateTime) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *DateTime) String() string {
	if d == nil || d.IsZero() {
		return zeroTimeValue
	} else {
		return d.Format(xtime.Fmt)
	}
}

func DateTimeNow() *DateTime {

	return &DateTime{Time: time.Now()}
}

func DateTimeFrom(from string) (*DateTime, error) {
	return dateTimeFrom(from, false)
}

func DateTimeStrictFrom(from string) (*DateTime, error) {
	return dateTimeFrom(from, true)
}

func DateTimesFrom(x, y string) (*DateTime, *DateTime, error) {
	return dateTimesFrom(x, y, false)
}

func DateTimesStrictFrom(x, y string) (*DateTime, *DateTime, error) {
	return dateTimesFrom(x, y, true)
}

func dateTimesFrom(x, y string, strict bool) (xdate *DateTime, ydate *DateTime, err error) {
	xdate, err = dateTimeFrom(x, strict)
	ydate, yerr := dateTimeFrom(y, strict)
	if err == nil {
		err = yerr
	}
	return xdate, ydate, err
}

func dateTimeFrom(from string, strict bool) (*DateTime, error) {
	if from == "" || from == zeroTimeValue {
		if strict {
			return nil, errors.WithStack(ErrDateTimeIsEmpty)
		} else {
			return nil, nil
		}
	}
	t, err := time.ParseInLocation(xtime.Fmt, from, time.Local)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &DateTime{Time: t}, nil
}

type SortDateTime []*DateTime

func (s SortDateTime) Len() int {
	return len(s)
}

func (s SortDateTime) Less(i, j int) bool {
	return (s[i]).Time.Before(s[j].Time)
}

func (s SortDateTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
