package xdb

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type orderTable struct {
	Model
	RoomNumber int64  `sql-col:"room_number"`
	Answer     string `sql-col:"answer"`
}

func (t *orderTable) TableName() string {
	return "order_table"
}

type orderTableRepo struct {
	Dao
}

func newOrderTableRepo(db *gorm.DB) *orderTableRepo {
	return &orderTableRepo{Dao: Dao{
		DB: db,
	}}
}

func (r *orderTableRepo) Save(ctx context.Context, record *orderTable) error {
	err := Trace(ctx, r.DB).Table((record.TableName())).Create(record).Error
	return errors.WithStack(err)
}

func (r *orderTableRepo) Update(ctx context.Context, record *orderTable) error {
	err := Trace(ctx, r.DB).Table(record.TableName()).Save(record).Error
	return errors.WithStack(err)
}

func (r *orderTableRepo) LoadByID(ctx context.Context, ID int64) (*orderTable, error) {
	result := &orderTable{}
	err := Trace(ctx, r.DB).Table(result.TableName()).Where("id = ?", ID).First(result).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}
