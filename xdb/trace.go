package xdb

import (
	"gitlab.intsig.net/cs-server2/kit/xlog"
	"gitlab.intsig.net/cs-server2/kit/xtrace"
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
	"unicode"
)

const (
	parentSpanKey = "opentracingParentSpan"
	spanKey       = "opentracingSpan"
	ctxKey        = "ctx"
)

// Trace sets span to gorm settings, returns cloned DB
func Trace(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx == nil {
		return db
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return db.InstantSet(ctxKey, ctx)
	}
	return db.Set(parentSpanKey, parentSpan).InstantSet(ctxKey, ctx)
}

// RegisterTraceCallbacks adds callbacks for tracing, you should call Trace to make them work
func RegisterTraceCallbacks(db *gorm.DB) {
	callbacks := newCallbacks()
	doRegisterCallbacks(db, "create", callbacks)
	doRegisterCallbacks(db, "query", callbacks)
	doRegisterCallbacks(db, "update", callbacks)
	doRegisterCallbacks(db, "delete", callbacks)
	doRegisterCallbacks(db, "row_query", callbacks)
}

type callbacks struct{}

func newCallbacks() *callbacks {
	return &callbacks{}
}

func (c *callbacks) beforeCreate(scope *gorm.Scope) {
	c.before(scope)
}
func (c *callbacks) afterCreate(scope *gorm.Scope) {
	c.after(scope, "INSERT")
}
func (c *callbacks) beforeQuery(scope *gorm.Scope) {
	c.before(scope)
}
func (c *callbacks) afterQuery(scope *gorm.Scope) {
	c.after(scope, "SELECT")
}
func (c *callbacks) beforeUpdate(scope *gorm.Scope) {
	c.before(scope)
}
func (c *callbacks) afterUpdate(scope *gorm.Scope) {
	c.after(scope, "UPDATE")
}
func (c *callbacks) beforeDelete(scope *gorm.Scope) {
	c.before(scope)
}
func (c *callbacks) afterDelete(scope *gorm.Scope) {
	c.after(scope, "DELETE")
}
func (c *callbacks) beforeRowQuery(scope *gorm.Scope) {
	c.before(scope)
}
func (c *callbacks) afterRowQuery(scope *gorm.Scope) {
	c.after(scope, "")
}

func (c *callbacks) before(scope *gorm.Scope) {
	val, ok := scope.Get(parentSpanKey)
	if !ok {
		return
	}
	parentSpan := val.(opentracing.Span)
	tr := parentSpan.Tracer()
	sp := tr.StartSpan("mysql", opentracing.ChildOf(parentSpan.Context()))
	//ext.DBType.Set(sp, "sql")

	val, ok = scope.Get(ctxKey)
	if ok {
		ctx := val.(context.Context)
		xlog.L(ctx).Check(zap.DebugLevel, "DB Exec").Write(
			zap.String("Start", "Sql"),
		)
	}

	scope.Set(spanKey, sp)
}

func (c *callbacks) after(scope *gorm.Scope, operation string) {
	if operation == "" {
		operation = strings.ToUpper(strings.Split(scope.SQL, " ")[0])
	}

	catchErr := scope.HasError() && scope.DB().Error != gorm.ErrRecordNotFound

	sqlVars := formattedValues(scope.SQLVars)
	val, ok := scope.Get(ctxKey)
	if ok {
		ctx := val.(context.Context)
		level := zap.InfoLevel
		if catchErr {
			level = zap.ErrorLevel
		} else if operation == "SELECT" {
			level = zap.DebugLevel
		}

		errF := zap.Skip()
		if catchErr {
			errF = zap.Error(scope.DB().Error)
		}
		xlog.L(ctx).Check(level, "DB Exec").Write(
			zap.String("op", operation),
			zap.String("statement", scope.SQL),
			zap.Strings("vars", sqlVars),
			zap.Int64("count", scope.DB().RowsAffected),
			errF,
		)
	}

	val, ok = scope.Get(spanKey)
	if ok {
		sp := val.(opentracing.Span)
		if catchErr {
			ext.Error.Set(sp, true)
		} else {
			ext.Error.Set(sp, false)
		}
		ext.DBType.Set(sp, "sql")
		ext.DBStatement.Set(sp, scope.SQL)
		sp.SetTag("db.op", operation)
		sp.SetTag("db.count", scope.DB().RowsAffected)
		sp.SetTag("db.vars", sqlVars)
		sp.Finish()
	}

}

func doRegisterCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	callbackName := fmt.Sprintf("gorm:%v", name)
	// gorm does some magic, if you pass CallbackProcessor here - nothing works
	switch name {
	case "create":
		db.Callback().Create().Before(callbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(callbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(callbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(callbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(callbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(callbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(callbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(callbackName).Register(afterName, c.afterDelete)
	case "row_query":
		db.Callback().RowQuery().Before(callbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().RowQuery().After(callbackName).Register(afterName, c.afterRowQuery)
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func formattedValues(values []interface{}) (formatted []string) {
	for _, rawValue := range values {
		indirectValue := reflect.Indirect(reflect.ValueOf(rawValue))
		if indirectValue.IsValid() {
			value := indirectValue.Interface()
			if t, ok := value.(time.Time); ok {
				formatted = append(formatted, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
			} else if b, ok := value.([]byte); ok {
				if str := string(b); isPrintable(str) {
					formatted = append(formatted, fmt.Sprintf("'%v'", str))
				} else {
					formatted = append(formatted, "'<binary>'")
				}
			} else if r, ok := value.(driver.Valuer); ok {
				if value, err := r.Value(); err == nil && value != nil {
					formatted = append(formatted, fmt.Sprintf("'%v'", value))
				} else {
					formatted = append(formatted, "NULL")
				}
			} else if e, ok := rawValue.(error); ok {
				formatted = append(formatted, fmt.Sprintf("err:%s", e.Error()))
			} else {
				switch value.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
					formatted = append(formatted, fmt.Sprintf("%v", value))
				default:
					formatted = append(formatted, fmt.Sprintf("'%v'", value))
				}
			}
		} else {
			formatted = append(formatted, "NULL")
		}
	}

	return formatted
}

// TableName 获取表明
// 支持环境隔离(BaggageVersion)
func TableName(db *gorm.DB, defaultName string) string {
	val, ok := db.Get(ctxKey)
	if !ok {
		return defaultName
	}

	ctx, ok := val.(context.Context)
	if !ok {
		return defaultName
	}

	version := getBaggageVersion(ctx)

	if version == "" {
		return defaultName
	}

	versionTable := fmt.Sprintf("%s_%s", defaultName, strings.ReplaceAll(version, "-", "_"))
	if db.HasTable(versionTable) {
		return versionTable
	}

	if db.Error != nil {
		xlog.S(ctx).Errorw("查询表是否存在报错", "table", versionTable, "err", errors.WithStack(db.Error))
	}

	return defaultName
}

func getBaggageVersion(ctx context.Context) string {
	meta := metautils.ExtractIncoming(ctx)
	version := meta.Get(xtrace.BaggageVersion)
	if version != "" {
		return version
	}

	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return parentSpan.BaggageItem(xtrace.BaggageVersion)
	}

	return ""
}
