package v2

import (
	"context"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/yituoshiniao/kit/xdb/v1"
	"github.com/yituoshiniao/kit/xlog"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

var defaultDbOptions = dbOptions{
	isMetrics: false,
}

type dbOptions struct {
	isMetrics bool
}

type DbOption func(*dbOptions)

func WithDBMetrics(isMetrics bool) DbOption {
	return func(o *dbOptions) {
		o.isMetrics = isMetrics
	}
}

func NewDb(conf v1.Config, zapLogger *zap.Logger, opts ...DbOption) (db *gorm.DB, fn func()) {
	o := &defaultDbOptions
	for _, opt := range opts {
		opt(o)
	}

	logger := zapgorm2.New(zapLogger)
	logger.LogLevel = glogger.Info //开启log记录
	logger.SetAsDefault()          // optional: configure gorm to use this zapgorm.Logger for callbacks

	//dsn := fmt.Sprintf(
	//	"%s:%s@(%s)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=%t&timeout=%ds",
	//)
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	var err error
	db, err = gorm.Open(mysql.Open(conf.Dsn), &gorm.Config{
		Logger: glogger.Discard, //禁用日志输出
		//Logger:                 logger, //设置log 使用zaplog
		SkipDefaultTransaction: true,
	})
	if err != nil {
		xlog.S(context.Background()).DPanicw("数据库连接错误", "err", err, "conf", conf)
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		xlog.S(context.Background()).DPanicw("数据库错误sqlDB", "err", err)
		panic(err)
	}

	err = db.Use(
		New(
			WithSqlParameters(true),
			WithMetrics(o.isMetrics),
		),
	)

	if err != nil {
		panic(err)
	}

	//err = db.Use(
	//	prometheus.New(prometheus.Config{
	//		DBName:          conf.DBName,   // 使用 `DBName` 作为指标 label
	//		RefreshInterval: conf.RefreshInterval,             // 指标刷新频率（默认为 15 秒）
	//		PushAddr:        conf.PushAddr, // 如果配置了 `PushAddr`，则推送指标
	//		MetricsCollector: []prometheus.MetricsCollector{
	//			&prometheus.MySQL{
	//				VariableNames: []string{"Threads_running"},
	//			},
	//		}, // 用户自定义指标
	//	}),
	//)

	sqlDB.SetMaxIdleConns(conf.MaxIdle)
	sqlDB.SetMaxOpenConns(conf.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.MaxLifetime) * time.Second)
	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			panic(errors.WithStack(err))
		}
	}
	return db, cleanup
}
