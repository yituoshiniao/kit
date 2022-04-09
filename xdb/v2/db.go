package v2

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	v1 "gitlab.intsig.net/cs-server2/kit/xdb/v1"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	//glogger "gorm.io/gorm/logger"
	glogger "gorm.io/gorm/logger"
	"time"
)

func NewDb(conf v1.Config, tracer opentracing.Tracer) (db *gorm.DB, fn func()) {
	//dsn := fmt.Sprintf(
	//	"%s:%s@(%s)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=%t&timeout=%ds",
	//)
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	var err error
	db, err = gorm.Open(mysql.Open(conf.Dsn), &gorm.Config{
		Logger: glogger.Discard,
		//SkipDefaultTransaction: false,
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	err = db.Use(
		New(
			WithTracer(nil),         //链路追踪
			WithLogResult(true),     // 是否记录查询结果
			WithSqlParameters(true), // sql 参数绑定解析
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
