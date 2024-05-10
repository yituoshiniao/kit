package xtask

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/yituoshiniao/kit/xlog"
	"github.com/yituoshiniao/kit/xrds"
)

// NewAsynqClient 任务队列queue 客户端
func NewAsynqClient(conf xrds.Config) (client *asynq.Client, cleanup func()) {
	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	cleanup = func() {
		xlog.S(context.Background()).Infow("NewAsynqClient-应用退出")
		err := client.Close()
		if err != nil {
			xlog.S(context.Background()).Infow("NewAsynqClient-应用退出err", "err", err)
		}
	}
	return client, cleanup
}
