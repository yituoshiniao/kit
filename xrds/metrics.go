package xrds

import (
	"context"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.intsig.net/cs-server2/kit/xlog"
)

var RdsAPICounter *kitprometheus.Counter
var RdsAPIHitsCounter *kitprometheus.Counter
var RdsAPIMissesCounter *kitprometheus.Counter

const (
	RdsAPICounterOperation    string = "operation"
	RdsAPICounterOperationGet string = "get"
)

func InitRdsAPICounterMetrics() {
	RdsAPICounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "redis",
			Name:      "operation_count",
			Help:      "operation  count of Counter metrics",
		},
		[]string{
			RdsAPICounterOperation,
		})

	RdsAPIHitsCounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "",
			Name:      "redis_hits_count",
			Help:      "get 缓存命中数量; redis hits count of Counter metrics",
		},
		[]string{RdsAPICounterOperation})

	RdsAPIMissesCounter = kitprometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Namespace: "",
			Name:      "redis_misses_count",
			Help:      "get 缓存未命中数量; redis misses count of Counter metrics",
		},
		[]string{RdsAPICounterOperation})

}

// processMetrics 原生 process 包装器，增加 metrics 埋点功能.
func processMetrics(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						xlog.S(ctx).Errorw("processMetrics  错误", "err", err)
					}
				}()
				RdsAPICounter.With(RdsAPICounterOperation, cmd.Name()).Add(1)
			}()

			obj := oldProcess(cmd)
			if obj != nil && obj == redis.Nil && cmd.Name() == RdsAPICounterOperationGet {
				RdsAPIMissesCounter.With(RdsAPICounterOperation, RdsAPICounterOperationGet).Add(1)
			}
			if obj == nil && (cmd.Name() == RdsAPICounterOperationGet) {
				RdsAPIHitsCounter.With(RdsAPICounterOperation, RdsAPICounterOperationGet).Add(1)
			}
			return obj
		}
	}
}

// processMetricsPipeline 原生 processPipeline 包装器，增加metrics埋点功能.
func processMetricsPipeline(ctx context.Context) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						xlog.S(ctx).Errorw("processMetricsPipeline 错误", "err", err)
					}
				}()
				RdsAPICounter.With(RdsAPICounterOperation, "processMetricsPipeline").Add(1)
			}()
			return oldProcess(cmds)
		}
	}
}
