package toot

import "time"

type QueueItem struct {
	ID         uint
	Content    string
	Media      []string
	SendStatus int
	SendResult string
	SendAt     time.Time
}

const (
	SENDSTATUS_WAITING = iota
	SENDSTATUS_SUCCESS
	SENDSTATUS_ERROR
)

const DATA_DIR = "media"
