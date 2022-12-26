package utils

import (
	"errors"

	"gorm.io/gorm"
)

func IgnoreErrNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
