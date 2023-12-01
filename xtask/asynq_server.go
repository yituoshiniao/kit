package xtask

import (
	"context"

	"github.com/hibiken/asynq"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xrds"
)

// AsynqQueue 队列名字
type AsynqQueue string

const (
	CriticalQueue AsynqQueue = "critical"
	DefaultQueue  AsynqQueue = "default"
	LowQueue      AsynqQueue = "low"
	HighQueue     AsynqQueue = "high"
	AdVideoQueue  AsynqQueue = "ad_video"
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

func NewAsynqServer(ctx context.Context, conf xrds.Config) (client *asynq.Server) {
	client = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.DB,
		},
		asynq.Config{
			// 并发执行的worker数量, 默认启动的worker数量是服务器的CPU个数。
			//Concurrency: 5,
			// 队列优先级
			Queues: map[string]int{
				string(CriticalQueue): 50,
				string(HighQueue):     20,
				string(DefaultQueue):  15,
				string(AdVideoQueue):  14,
				string(LowQueue):      1,
			},
			// 错误回调处理
			ErrorHandler: asynq.ErrorHandlerFunc(reportError),
			Logger:       NewLogger(),
		},
	)
	return client
}

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(args ...interface{}) {
	xlog.S(context.Background()).Info(args)
}

func (l *Logger) Debug(args ...interface{}) {
	xlog.S(context.Background()).Debug(args)
}

func (l *Logger) Warn(args ...interface{}) {
	xlog.S(context.Background()).Warn(args)
}

func (l *Logger) Error(args ...interface{}) {
	xlog.S(context.Background()).Error(args)
}

func (l *Logger) Fatal(args ...interface{}) {
	xlog.S(context.Background()).Fatal(args)
}

// reportError 错误回调处理函数
func reportError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	if retried >= maxRetry {
		//一旦一个任务耗尽它的重试次数，任务将转为Archived状态
		xlog.S(ctx).Errorw("重试次数已用完", "err", err, "task", string(task.Payload()))
	}

	xlog.S(ctx).Errorw("任务执行错误", "err", err, "task", string(task.Payload()), "重试次数", retried)
}
