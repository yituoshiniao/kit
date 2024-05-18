package v2

import (
	"context"
	"strings"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/pkg/errors"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"

	"github.com/yituoshiniao/kit/xlog"
)

var DBAPICounter *kitprometheus.Counter
var DBErrCounter *kitprometheus.Counter

const (
	DBTable     string = "table"
	DBOperation string = "operation"
	DBSQL       string = "dbSql"
	DBErrMsg    string = "errMsg"
)

func InitDBCounterMetrics() {
	DBAPICounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "mysql",
			Name:      "api_count",
			Help:      "db count of Counter metrics",
		},
		[]string{
			DBTable,
			DBOperation,
		},
	)
	DBErrCounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "mysql",
			Name:      "err_count",
			Help:      "db err count of Counter metrics",
		},
		[]string{
			DBTable,
			DBOperation,
			DBSQL,
			DBErrMsg,
		},
	)
}

func (p opentracingPlugin) metrics(db *gorm.DB) {
	// 错误统计
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					xlog.S(context.Background()).Errorw("metricsAfter ErrRecordNotFound 错误", "err", errors.WithStack(errors.Errorf("%v", err)))
				}
			}()

			DBErrCounter.With(
				DBTable, db.Statement.Table,
				DBOperation, strings.ToUpper(strings.Split(db.Statement.SQL.String(), " ")[0]),
				DBErrMsg, db.Error.Error(),
				DBSQL, db.Statement.SQL.String(),
			).Add(1)
		}()
	}

	// 操作统计
	DBAPICounter.With(
		DBTable, db.Statement.Table,
		DBOperation, strings.ToUpper(strings.Split(db.Statement.SQL.String(), " ")[0]),
	).Add(1)

}
