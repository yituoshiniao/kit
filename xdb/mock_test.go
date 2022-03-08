package xdb

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/pkg/errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetInsertStmtFromEntity(t *testing.T) {
	t.Parallel()
	testTable := &orderTable{}
	insertExpression := GetInsertStmtFromEntity(testTable, testTable.TableName())
	assert.Equal(t, "INSERT INTO `order_table` (`room_number`,`answer`) VALUES (?,?)", insertExpression)
}

func TestGetUpdateStmtFromEntity(t *testing.T) {
	t.Parallel()
	testTable := &orderTable{}
	updateExpression := GetUpdateStmtFromEntity(testTable, testTable.TableName(), "id")
	assert.Equal(t, "UPDATE `order_table` SET `room_number` = ?, `answer` = ? WHERE `order_table`.`id` = ?", updateExpression)
}

func TestGormMock(t *testing.T) {
	scenarios := []struct {
		desc   string
		mockFn func() *GormMock
		testFn func(t *testing.T, gormMock *GormMock)
	}{
		{
			desc: "Save - Success",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				testRecord := getDummyTestOrderTableEntity()
				insertExpression := GetInsertStmtFromEntity(testRecord, testRecord.TableName())
				testGormMock.ExpectExecWithResult(sqlmock.NewResult(1, 1), insertExpression, testRecord.RoomNumber, testRecord.Answer)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntity()

				err := testOrderTableRepo.Save(context.Background(), testRecord)
				assert.Nil(t, err)
			},
		},
		{
			desc: "Save - Failure",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				testRecord := getDummyTestOrderTableEntity()
				insertExpression := GetInsertStmtFromEntity(testRecord, testRecord.TableName())
				testGormMock.ExpectExecWithError(assert.AnError, insertExpression, testRecord.RoomNumber, testRecord.Answer)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntity()

				err := testOrderTableRepo.Save(context.Background(), testRecord)
				assert.Equal(t, errors.WithStack(assert.AnError).Error(), err.Error())
			},
		},
		{
			desc: "Update - Success",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				testRecord := getDummyTestOrderTableEntity()
				testRecord.ID = 1

				updateExpression := GetUpdateStmtFromEntity(testRecord, testRecord.TableName(), "id")
				testGormMock.ExpectExecWithResult(sqlmock.NewResult(0, 1), updateExpression, testRecord.RoomNumber, testRecord.Answer, testRecord.ID)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntity()
				testRecord.ID = 1

				err := testOrderTableRepo.Update(context.Background(), testRecord)
				assert.Nil(t, err)
			},
		},
		{
			desc: "Update - Failure",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				testRecord := getDummyTestOrderTableEntity()
				testRecord.ID = 1

				updateExpression := GetUpdateStmtFromEntity(testRecord, testRecord.TableName(), "id")
				testGormMock.ExpectExecWithError(assert.AnError, updateExpression, testRecord.RoomNumber, testRecord.Answer, testRecord.ID)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntity()
				testRecord.ID = 1

				err := testOrderTableRepo.Update(context.Background(), testRecord)
				assert.Equal(t, errors.WithStack(assert.AnError).Error(), err.Error())
			},
		},
		{
			desc: "LoadByID - Success",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				loadExpression := "SELECT * FROM `order_table` WHERE (id = ?) ORDER BY `order_table`.`id` ASC LIMIT 1"
				testGormMock.ExpectQueryWithResult(getTestOrderTableMockSQLRows(getDummyTestOrderTableEntityWithID()), loadExpression, 1)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntityWithID()

				result, err := testOrderTableRepo.LoadByID(context.Background(), 1)
				assert.Nil(t, err)
				assert.Equal(t, testRecord, result)
			},
		},
		{
			desc: "LoadByID - Success(Used Models2SQLMockRows)",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				loadExpression := "SELECT * FROM `order_table` WHERE (id = ?) ORDER BY `order_table`.`id` ASC LIMIT 1"
				testGormMock.ExpectQueryWithResult(testGormMock.Models2SQLMockRows(getDummyTestOrderTableEntityWithID()), loadExpression, 1)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())
				testRecord := getDummyTestOrderTableEntityWithID()

				result, err := testOrderTableRepo.LoadByID(context.Background(), 1)
				assert.Nil(t, err)
				assert.Equal(t, testRecord, result)
			},
		},
		{
			desc: "LoadByID - Failure",
			mockFn: func() *GormMock {
				testGormMock := NewGormMock()
				loadExpression := "SELECT * FROM `order_table` WHERE (id = ?) ORDER BY `order_table`.`id` ASC LIMIT 1"
				testGormMock.ExpectQueryWithError(assert.AnError, loadExpression, 1)
				return testGormMock
			},
			testFn: func(t *testing.T, gormMock *GormMock) {
				testOrderTableRepo := newOrderTableRepo(gormMock.GormDB())

				result, err := testOrderTableRepo.LoadByID(context.Background(), 1)
				assert.Equal(t, errors.WithStack(assert.AnError).Error(), err.Error())
				assert.Nil(t, result)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.desc, func(t *testing.T) {
			gormMock := scenario.mockFn()
			scenario.testFn(t, gormMock)
			assert.NoError(t, gormMock.mock.ExpectationsWereMet())
			gormMock.Close()
		})
	}
}

func getDummyTestOrderTableEntity() *orderTable {
	return &orderTable{
		RoomNumber: 123232,
		Answer:     "answer",
	}
}

func getDummyTestOrderTableEntityWithID() *orderTable {
	return &orderTable{
		Model:      Model{ID: 1},
		RoomNumber: 123232,
		Answer:     "answer",
	}
}

func getTestOrderTableMockSQLRows(expects ...*orderTable) *sqlmock.Rows {
	columns := []string{`id`, `room_number`, `answer`}
	sqlmockRows := sqlmock.NewRows(columns)
	for _, expect := range expects {
		sqlmockRows.AddRow([]driver.Value{
			expect.ID,
			expect.RoomNumber,
			expect.Answer,
		}...)
	}
	return sqlmockRows
}
