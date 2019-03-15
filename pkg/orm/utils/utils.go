package utils

import (
	"github.com/jinzhu/gorm"
)

//CheckExists is method for checking existing record in DB
func CheckExists(db *gorm.DB, object interface{}, id interface{}) (bool, error) {
	count := 0
	err := db.Model(object).Where("id = ?", id).Limit(1).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}