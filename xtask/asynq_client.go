package xtask

import (
	"context"

	"github.com/hibiken/asynq"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xrds"
)

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
