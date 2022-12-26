package model

type Config struct {
	Name string `gorm:"primaryKey"`
	Data string
}
