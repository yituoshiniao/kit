package xtask

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xrds"
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

// NewAsynqScheduler asynq 调度任务、和定时任务类似
func NewAsynqScheduler(ctx context.Context, conf xrds.Config) (client *asynq.Scheduler) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	client = asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.DB,
		},

		&asynq.SchedulerOpts{
			Location:        loc,
			PostEnqueueFunc: handlePostEnqueueFunc,
			PreEnqueueFunc:  handlePreEnqueueFunc,
		},
	)
	return client
}

// handlePreEnqueueFunc 调度器将任务排入队列之前调用它
func handlePreEnqueueFunc(task *asynq.Task, opts []asynq.Option) {
	xlog.S(context.Background()).Infow("handlePreEnqueueFunc-处理", "opts",
		opts, "task.payload", string(task.Payload()))

}

// handlePostEnqueueFunc 调度程序将任务排入队列后调用的
func handlePostEnqueueFunc(taskInfo *asynq.TaskInfo, err error) {
	xlog.S(context.Background()).Infow("handlePostEnqueueFunc-处理", "taskInfo", taskInfo, "payload", string(taskInfo.Payload))
	if err != nil {
		xlog.S(context.Background()).Errorw("handlePostEnqueueFunc-处理err", "err", err)
	}
}
