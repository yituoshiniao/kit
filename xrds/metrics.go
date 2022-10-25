package xrds

import (
	"context"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.intsig.net/cs-server2/kit/xlog"
)

var RdsAPICounter *kitprometheus.Counter

const (
	RdsAPICounterOperation string = "operation"
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
			return oldProcess(cmd)
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
