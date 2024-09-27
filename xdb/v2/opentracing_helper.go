package v2

import (
	"context"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/yituoshiniao/kit/xlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	_prefix      = "gorm.opentelemetry"
	_errorTagKey = "error"
)

var (
	_tableTagKey        = keyWithPrefix("table")
	_resultLogKey       = keyWithPrefix("result")
	_sqlLogKey          = keyWithPrefix("sql")
	_rowsAffectedLogKey = keyWithPrefix("rowsAffected")
)

func keyWithPrefix(key string) string {
	return _prefix + "." + key
}

var (
	opentracingSpanKey = "opentracing:span"
	timeMsKey          = "timeMs"
	ctxKey             = "ctx"
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
)

func (p opentracingPlugin) injectBefore(db *gorm.DB, op operationName) {
	if db == nil {
		return
	}

	if db.Statement == nil || db.Statement.Context == nil {
		xlog.S(context.TODO()).Error("could not inject span from nil Statement.Context or nil Statement")
		return
	}

	tracer := otel.Tracer("gorm-tracer") // 创建 tracer
	ctx, sp := tracer.Start(db.Statement.Context, op.String())
	db.InstanceSet(timeMsKey, time.Now())
	db.InstanceSet(opentracingSpanKey, sp)
	db.InstanceSet(ctxKey, ctx)
}

func (p opentracingPlugin) extractAfter(db *gorm.DB) {
	if db == nil {
		xlog.S(context.TODO()).Debug("DB is nil 错误")
		return
	}
	if db.Statement == nil || db.Statement.Context == nil {
		xlog.S(context.TODO()).Error("could not extract span from nil Statement.Context or nil Statement")
		return
	}

	sTime, timeOk := db.InstanceGet(timeMsKey)
	v, okCtx := db.InstanceGet(ctxKey)
	if okCtx {
		ctx := v.(context.Context)
		logFields := appendLogSql(db, p.opt.logResult, p.opt.logSqlParameters)
		if timeOk {
			logFields = append(logFields, zap.String("gorm耗时", time.Now().Sub(sTime.(time.Time)).String()))
			xlog.L(ctx).Info("[Gorm]:Exec", logFields...)
		} else {
			xlog.L(ctx).Info("[Gorm]:Exec", logFields...)
		}
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			xlog.S(ctx).Errorw("gorm 错误信息", "err", db.Error)
		}
	}

	v, ok := db.InstanceGet(opentracingSpanKey)
	if !ok || v == nil {
		xlog.S(context.TODO()).Debug("InstanceGet opentracingSpanKey 错误")
		return
	}

	sp, ok := v.(trace.Span)
	if !ok || sp == nil {
		xlog.S(context.TODO()).Debug("v.(trace.Span) 错误")
		return
	}
	defer sp.End() // 结束跨度

	tag(sp, db, p.opt.errorTagHook)
	log(sp, db, p.opt.logResult, p.opt.logSqlParameters)
}

// errorTagHook will be called while gorm.DB got an error and we need a way to mark this error
type errorTagHook func(sp trace.Span, err error)

func defaultErrorTagHook(sp trace.Span, err error) {
	sp.SetAttributes(attribute.Bool(_errorTagKey, true))
}

func tag(sp trace.Span, db *gorm.DB, errorTagHook errorTagHook) {
	if err := db.Error; err != nil && nil != errorTagHook {
		errorTagHook(sp, err)
	}

	sp.SetAttributes(attribute.String(_tableTagKey, db.Statement.Table)) // 设置标签
}

func log(sp trace.Span, db *gorm.DB, verbose bool, logSqlVariables bool) {
	if logSqlVariables {
		sp.AddEvent("SQL executed", trace.WithAttributes(attribute.String(_sqlLogKey, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...))))
	} else {
		sp.AddEvent("SQL executed", trace.WithAttributes(attribute.String(_sqlLogKey, db.Statement.SQL.String())))
	}

	sp.SetAttributes(attribute.Int64(_rowsAffectedLogKey, db.Statement.RowsAffected))

	if err := db.Error; err != nil {
		sp.RecordError(err)
		sp.SetAttributes(attribute.Bool("error", true))

	}
}

func appendLogSql(db *gorm.DB, verbose bool, logSqlVariables bool) (logField []zap.Field) {
	logField = make([]zap.Field, 0)
	if logSqlVariables {
		logField = append(logField, zap.String(_sqlLogKey, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
	} else {
		logField = append(logField, zap.String(_sqlLogKey, db.Statement.SQL.String()))
	}

	logField = append(logField, zap.Int64(_rowsAffectedLogKey, db.Statement.RowsAffected))
	// logField = append(logField, zap.Reflect("Vars", db.Statement.Vars))
	logField = append(logField, zap.String("operation", strings.ToUpper(strings.Split(db.Statement.SQL.String(), " ")[0])))

	// 操作类型  SELECT, DELETE , CREATE, ALTER, INSET 等
	// fields["Schema.Table"] = db.Statement.Schema.Table

	// 记录返回数据
	// if verbose && db.Statement.Dest != nil {
	//	// DONE(@yeqown) fill result fields into span log
	//	// FIXED(@yeqown) db.Statement.Dest still be metatable now ?
	//	v, err := json.Marshal(db.Statement.Dest)
	//	if err == nil {
	//		fields[_resultLogKey] = *(*string)(unsafe.Pointer(&v))
	//	} else {
	//		db.Logger.Error(context.Background(), "could not marshal db.Statement.Dest: %v", err)
	//	}
	// }

	return
}
