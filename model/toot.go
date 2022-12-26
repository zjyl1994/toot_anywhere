package model

import (
	"time"

	"gorm.io/gorm"
)

type Toot struct {
	gorm.Model
	TootContent string
	MediaFiles  string
	SendAfter   time.Time
	SendStatus  int
	SendResult  string
}
