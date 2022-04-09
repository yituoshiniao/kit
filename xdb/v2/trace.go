package v2

import (
	"context"
	"gorm.io/gorm"
)

// Trace sets span to gorm settings, returns cloned DB
func Trace(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx == nil {
		return db
	}
	db.Statement.Context = ctx // log 日志追踪
	return db
}
