package v2

// operationName defines a type to wrap the name of each operation name.
type operationName string

// String returns the actual string of operationName.
func (op operationName) String() string {
	return string(op)
}

const (
	_createOp operationName = "mysql.create"
	_updateOp operationName = "mysql.update"
	_queryOp  operationName = "mysql.query"
	_deleteOp operationName = "mysql.delete"
	_rowOp    operationName = "mysql.row"
	_rawOp    operationName = "mysql.raw"
)

// operationStage indicates the timing when the operation happens.
type operationStage string

// Name returns the actual string of operationStage.
func (op operationStage) Name() string {
	return string(op)
}

const (
	_stageBeforeCreate operationStage = "opentracing:before_create"
	_stageAfterCreate  operationStage = "opentracing:after_create"
	_stageBeforeUpdate operationStage = "opentracing:before_update"
	_stageAfterUpdate  operationStage = "opentracing:after_update"
	_stageBeforeQuery  operationStage = "opentracing:before_query"
	_stageAfterQuery   operationStage = "opentracing:after_query"
	_stageBeforeDelete operationStage = "opentracing:before_delete"
	_stageAfterDelete  operationStage = "opentracing:after_delete"
	_stageBeforeRow    operationStage = "opentracing:before_row"
	_stageAfterRow     operationStage = "opentracing:after_row"
	_stageBeforeRaw    operationStage = "opentracing:before_raw"
	_stageAfterRaw     operationStage = "opentracing:after_raw"
)
