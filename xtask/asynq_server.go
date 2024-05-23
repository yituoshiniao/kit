package xtask

import (
	"context"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"github.com/yituoshiniao/kit/xlog"
	"github.com/yituoshiniao/kit/xrds"
)

type AsynqQueue string

// 注意默认task 不指定队列 默认使用default,当指定的队列不在如下中,那么注册的handle无法处理任务
const (
	CriticalQueue       AsynqQueue = "critical"
	DefaultQueue        AsynqQueue = "default"
	LowQueue            AsynqQueue = "low"
	HighQueue           AsynqQueue = "high"
	AdVideoQueue        AsynqQueue = "adVideo"
	GroupQueue          AsynqQueue = "groupQueue"
	GroupSchedulerQueue AsynqQueue = "schedulerQuee"
)

/****
	以下是不同生命周期状态的列表
	Scheduled：任务正在等待未来处理（仅适用于具有ProcessAt或ProcessIn选项的任务）。
	Pending：任务已准备好进行处理，并将由一个空闲的工作器接收。
	Active：任务正在被工作器处理（即处理程序正在处理该任务）。
	Retry：工作器无法处理任务，任务正在等待将来重试。
	Archived：任务达到最大重试次数，并存储在归档中以供手动检查。
	Completed：任务已成功处理，并保留到保留时间到期为止（仅适用于具有Retention选项的任务）。
****/

// NewAsynqServer asynq 服务端
func NewAsynqServer(ctx context.Context, conf xrds.Config) (client *asynq.Server) {
	client = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.DB,
		},
		asynq.Config{
			// 并发执行的worker数量, 默认启动的worker数量是服务器的CPU个数。
			// Concurrency: 5,
			// 队列优先级
			Queues: map[string]int{
				string(CriticalQueue):       50,
				string(HighQueue):           10,
				string(GroupSchedulerQueue): 10,
				string(DefaultQueue):        15,
				string(AdVideoQueue):        9,
				string(GroupQueue):          5,
				string(LowQueue):            1,
			},
			// 错误回调处理
			ErrorHandler: asynq.ErrorHandlerFunc(reportError),
			Logger:       NewLogger(ctx),

			// 组聚合参数
			GroupAggregator:  asynq.GroupAggregatorFunc(aggregate),
			GroupGracePeriod: 10 * time.Second, // 组的优雅延迟时间, 每10秒聚合一次
			GroupMaxDelay:    30 * time.Second, // 组的最大延迟时间
			GroupMaxSize:     10,               // 组的最大尺寸
		},
	)
	return client
}

// AggregateTypeName 聚合 typename
const AggregateTypeName = "aggregated-task"

// 简单的聚合函数。
// 将所有任务的消息组合在一起，每个消息占一行。
func aggregate(group string, tasks []*asynq.Task) *asynq.Task {
	xlog.S(context.Background()).Infow("aggregate聚合信息", "group", group, "len", len(tasks))
	var b strings.Builder
	for _, t := range tasks {
		b.Write(t.Payload())
		b.WriteString("\n")
	}
	return asynq.NewTask(AggregateTypeName, []byte(b.String()))
	// return asynq.NewTask("email:groupDeliver", []byte(b.String()))
}

// reportError 错误回调处理函数
func reportError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	if retried >= maxRetry {
		// 一旦一个任务耗尽它的重试次数，任务将转为Archived状态
		xlog.S(ctx).Errorw("重试次数已用完", "err", err, "task", string(task.Payload()))
	}

	xlog.S(ctx).Errorw("任务执行错误", "err", err, "task", string(task.Payload()), "重试次数", retried)
}
