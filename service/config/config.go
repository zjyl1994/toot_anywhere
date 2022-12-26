package config

import (
	"errors"

	"github.com/zjyl1994/toot_anywhere/model"
	"github.com/zjyl1994/toot_anywhere/vars"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Get(key string) (string, error) {
	var m model.Config
	err := vars.DB.First(&m, "name = ?", key).Error
	if err != nil {
		return "", err
	}
	return m.Data, nil
}

func IsSet(key string) (bool, error) {
	if _, err := Get(key); err == nil {
		return true, nil
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
}

func MustGet(key, defautValue string) string {
	result, err := Get(key)
	if err != nil {
		return defautValue
	} else {
		return result
	}
}

func Set(key, value string) error {
	m := model.Config{
		Name: key,
		Data: value,
	}
	return vars.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"data"}),
	}).Create(&m).Error
}
