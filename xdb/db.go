//go:generate msgp -tests=false
//msgp:ignore Dao
package xdb

import (
	"context"
	"github.com/dlmiddlecote/sqlstats"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	gormbulk "github.com/t-tiger/gorm-bulk-insert/v2"
	"time"
)

func Open(mysql Mysql) (*gorm.DB, func()) {
	return New(mysql)
}

func New(conf Config) (*gorm.DB, func()) {
	if db, err := gorm.Open("mysql", conf.Dsn); err != nil {
		panic(errors.WithStack(err))
	} else {
		db.LogMode(conf.LogMode)
		db.DB().SetMaxIdleConns(conf.MaxIdle)
		db.DB().SetMaxOpenConns(conf.MaxOpen)
		db.DB().SetConnMaxLifetime(time.Duration(conf.MaxLifetime) * time.Second)

		RegisterTraceCallbacks(db)

		// Prometheus Register
		cfg, _ := mysql.ParseDSN(conf.Dsn)
		collector := sqlstats.NewStatsCollector(cfg.DBName, db.DB())
		// https://github.com/dlmiddlecote/sqlstats/commit/c437853fbcb30892e95d86b255d26d4967d99425
		_ = prometheus.Register(collector)

		cleanup := func() {
			if err := db.Close(); err != nil {
				panic(errors.WithStack(err))
			}
		}

		return db, cleanup
	}
}

type Model struct {
	ID int64 `gorm:"primary_key"`
}

type Dao struct {
	DB *gorm.DB
}

func (d *Dao) Create(ctx context.Context, entity interface{}) error {
	return d.CreateTx(ctx, d.DB, entity)
}

func (d *Dao) Save(ctx context.Context, entity interface{}) error {
	return d.SaveTx(ctx, d.DB, entity)
}

func (d *Dao) CreateTx(ctx context.Context, tx *gorm.DB, entity interface{}) error {
	return Trace(ctx, tx).Create(entity).Error
}

func (d *Dao) SaveTx(ctx context.Context, tx *gorm.DB, entity interface{}) error {
	err := Trace(ctx, tx).Save(entity).Error
	return errors.WithStack(err)
}

// UpdatesWithZeroValue 更新结构体，支持零值
//
// 警告：当使用 struct 更新时，GORM只会更新那些非零值的字段
// 对于下面的操作，不会发生任何更新，"", 0, false 都是其类型的零值
func UpdatesWithZeroValue(db *gorm.DB, values interface{}, ignoreProtectedAttrs ...bool) *gorm.DB {
	data := map[string]interface{}{}
	for _, field := range db.NewScope(values).Fields() {
		if field.IsIgnored {
			continue
		}
		data[field.DBName] = field.Field.Interface()
	}

	if len(data) == 0 {
		return db
	}

	return db.Updates(data, ignoreProtectedAttrs...)
}

// BulkCreate 批量创建记录
//
// [objects] must be a slice of struct.
//
// [chunkSize] is a number of variables embedded in query. To prevent the error which occurs embedding a large number of variables at once
// and exceeds the limit of prepared statement. Larger size normally leads to better performance, in most cases 2000 to 3000 is reasonable.
//
// [excludeColumns] is column names to exclude from insert.
func BulkCreate(db *gorm.DB, objects []interface{}, chunkSize int, excludeColumns ...string) error {
	err := gormbulk.BulkInsert(db, objects, chunkSize, excludeColumns...)
	return errors.WithStack(err)
}

