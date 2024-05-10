package v2

import "gorm.io/gorm"

func NotDelete(db *gorm.DB) *gorm.DB {
	return db.Where("is_del = 0")
}

func Delete(db *gorm.DB) *gorm.DB {
	return db.Where("is_del != 0")
}
