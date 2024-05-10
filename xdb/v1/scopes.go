package v1

import "github.com/jinzhu/gorm"

func NotDelete(db *gorm.DB) *gorm.DB {
	return db.Where("delete_time is null or delete_time=0")
}

func Delete(db *gorm.DB) *gorm.DB {
	return db.Where("delete_time is not null and delete_time != 0")
}
