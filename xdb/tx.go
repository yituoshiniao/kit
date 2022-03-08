package xdb

import (
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func Begin(ctx context.Context, db *gorm.DB) (tx *gorm.DB, err error) {
	tx = Trace(ctx, db).Begin()
	if err := tx.Error; err != nil {
		return tx, errors.Wrap(err, "开启事务失败")
	}

	return tx, nil
}

func Commit(ctx context.Context, tx *gorm.DB) error {
	if err := Trace(ctx, tx).Commit().Error; err != nil {
		return errors.Wrap(err, "提交事务失败")
	}
	return nil
}

func Rollback(ctx context.Context, tx *gorm.DB) {
	if err := Trace(ctx, tx).Rollback().Error; err != nil {
		xlog.S(ctx).Errorw("回滚事务失败", "err", err)
	}
}
