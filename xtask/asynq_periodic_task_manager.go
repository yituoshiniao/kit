package xtask

import (
	"context"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xrds"
	"gopkg.in/yaml.v2"
)

type AsynqConfig struct {
	Enable                  bool   `yaml:"enable" json:"enable"`
	Host                    string `yaml:"host" json:"host"`
	Port                    int    `yaml:"port" json:"port"`
	EnableRegisterSched     bool   `yaml:"enableRegisterSched" json:"enableRegisterSched"`
	PeriodicTaskConfig      string `yaml:"periodicTaskConfig" json:"periodicTaskConfig"`
	EnablePeriodicTaskSched bool   `yaml:"enablePeriodicTaskSched" json:"enablePeriodicTaskSched"`
}

func NewPeriodicTaskManager(ctx context.Context, conf xrds.Config, asynqConfig AsynqConfig) (client *asynq.PeriodicTaskManager) {
	redisConnOpt := asynq.RedisClientOpt{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	schedulerOpts := &asynq.SchedulerOpts{
		Location: loc,
	}

	provider := &FileBasedConfigProvider{filename: asynqConfig.PeriodicTaskConfig, enablePeriodicTaskSched: asynqConfig.EnablePeriodicTaskSched}
	client, err = asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               redisConnOpt,
			PeriodicTaskConfigProvider: provider,         // this provider object is the interface to your config source
			SyncInterval:               10 * time.Second, // this field specifies how often sync should happen

			SchedulerOpts: schedulerOpts,
		},
	)
	if err != nil {
		panic(err)
	}

	return client
}

// FileBasedConfigProvider implements asynq.PeriodicTaskConfigProvider interface.
type FileBasedConfigProvider struct {
	filename                string
	enablePeriodicTaskSched bool
}

// GetConfigs Parses the yaml file and return a list of PeriodicTaskConfigs.
func (p *FileBasedConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	data, err := os.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}
	var c PeriodicTaskConfigContainer
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	var configs []*asynq.PeriodicTaskConfig

	if !p.enablePeriodicTaskSched {
		xlog.S(context.Background()).Info("不需要注册动态调度任务")
		return configs, nil
	}

	//task 选项设置
	opts := []asynq.Option{
		asynq.Retention(time.Hour * 72),
		asynq.Queue(string(GroupSchedulerQueue)),
	}
	for _, cfg := range c.Configs {
		configs = append(
			configs, &asynq.PeriodicTaskConfig{
				Cronspec: cfg.Cronspec,
				Task: asynq.NewTask(
					cfg.TaskType,
					nil,
				),
				Opts: opts,
			},
		)
	}
	return configs, nil
}

type PeriodicTaskConfigContainer struct {
	Configs []*Config `yaml:"configs"`
}

type Config struct {
	Cronspec string `yaml:"cronspec"`
	TaskType string `yaml:"task_type"`
}
