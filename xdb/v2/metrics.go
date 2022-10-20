package v2

import (
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
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
