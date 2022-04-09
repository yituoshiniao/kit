package gormopentracing

import (
	"context"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"gorm.io/gorm"
)

const (
	_prefix      = "gorm.opentracing"
	_errorTagKey = "error"
)

var (
	// span.Tag keys
	_tableTagKey = keyWithPrefix("table")
	// span.Log keys
	//_errorLogKey        = keyWithPrefix("error")
	_resultLogKey       = keyWithPrefix("result")
	_sqlLogKey          = keyWithPrefix("sql")
	_rowsAffectedLogKey = keyWithPrefix("rowsAffected")
)

func keyWithPrefix(key string) string {
	return _prefix + "." + key
}

var (
	opentracingSpanKey = "opentracing:span"
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
)

func (p opentracingPlugin) injectBefore(db *gorm.DB, op operationName) {
	// make sure context could be used
	if db == nil {
		return
	}

	if db.Statement == nil || db.Statement.Context == nil {
		db.Logger.Error(context.TODO(), "could not inject sp from nil Statement.Context or nil Statement")
		return
	}

	sp, _ := opentracing.StartSpanFromContextWithTracer(db.Statement.Context, p.opt.tracer, op.String())
	db.InstanceSet(opentracingSpanKey, sp)
}

func (p opentracingPlugin) extractAfter(db *gorm.DB) {
	// make sure context could be used
	if db == nil {
		return
	}
	if db.Statement == nil || db.Statement.Context == nil {
		db.Logger.Error(context.TODO(), "could not extract sp from nil Statement.Context or nil Statement")
		return
	}

	// extract sp from db context
	//sp := opentracing.SpanFromContext(db.Statement.Context)
	v, ok := db.InstanceGet(opentracingSpanKey)
	if !ok || v == nil {
		return
	}

	sp, ok := v.(opentracing.Span)
	if !ok || sp == nil {
		return
	}
	defer sp.Finish()

	// tag and log fields we want.
	tag(sp, db, p.opt.errorTagHook)
	log(sp, db, p.opt.logResult, p.opt.logSqlParameters)
}

// errorTagHook will be called while gorm.DB got an error and we need a way to mark this error
// in current opentracing.Span. Of course, you can use sp.LogField in this hook, but it's not
// recommended to.
//
// mark an error tag in sp as default:
//
// sp.SetTag(sp.SetTag(_errorTagKey, true))
type errorTagHook func(sp opentracing.Span, err error)

func defaultErrorTagHook(sp opentracing.Span, err error) {
	sp.SetTag(_errorTagKey, true)
}

// tag called after operation
func tag(sp opentracing.Span, db *gorm.DB, errorTagHook errorTagHook) {
	if err := db.Error; err != nil && nil != errorTagHook {
		errorTagHook(sp, err)
		// sp.SetTag(_errorTagKey, true)
	}

	sp.SetTag(_tableTagKey, db.Statement.Table)
}

// log called after operation
func log(sp opentracing.Span, db *gorm.DB, verbose bool, logSqlVariables bool) {
	fields := make([]opentracinglog.Field, 0, 4)
	fields = appendSql(fields, db, logSqlVariables)
	fields = append(fields, opentracinglog.Object(_rowsAffectedLogKey, db.Statement.RowsAffected))

	// log error
	if err := db.Error; err != nil {
		fields = append(fields, opentracinglog.Error(err))
	}

	if verbose && db.Statement.Dest != nil {
		// DONE(@yeqown) fill result fields into span log
		// FIXED(@yeqown) db.Statement.Dest still be metatable now ?
		v, err := json.Marshal(db.Statement.Dest)
		if err == nil {
			fields = append(fields, opentracinglog.String(_resultLogKey, *(*string)(unsafe.Pointer(&v))))
		} else {
			db.Logger.Error(context.Background(), "could not marshal db.Statement.Dest: %v", err)
		}
	}

	sp.LogFields(fields...)
}

func appendSql(fields []opentracinglog.Field, db *gorm.DB, logSqlVariables bool) []opentracinglog.Field {
	if logSqlVariables {
		fields = append(fields, opentracinglog.String(_sqlLogKey,
			db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
	} else {
		fields = append(fields, opentracinglog.String(_sqlLogKey, db.Statement.SQL.String()))
	}
	return fields
}
