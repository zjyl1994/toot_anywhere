package toot

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/zjyl1994/toot_anywhere/model"
	"github.com/zjyl1994/toot_anywhere/utils"
	"github.com/zjyl1994/toot_anywhere/vars"
	"gorm.io/gorm"
)

func PushQueue(content string, media []string, delaySecond int) error {
	m := model.Toot{
		TootContent: content,
		MediaFiles:  strings.Join(media, "\n"),
	}
	m.SendStatus = SENDSTATUS_WAITING
	m.SendAfter = time.Now().Add(time.Duration(delaySecond) * time.Second)
	return vars.DB.Create(&m).Error
}

func ListQueue() ([]QueueItem, error) {
	var toots []model.Toot
	err := vars.DB.Order("send_after desc").Find(&toots).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []QueueItem{}, nil
		}
		return nil, err
	}
	return lo.Map(toots, func(item model.Toot, _ int) QueueItem {
		var mediaFiles []string
		if len(strings.TrimSpace(item.MediaFiles)) > 0 {
			mediaFiles = strings.Split(item.MediaFiles, "\n")
		}

		return QueueItem{
			ID:         item.ID,
			Content:    item.TootContent,
			Media:      mediaFiles,
			SendStatus: item.SendStatus,
			SendResult: item.SendResult,
			SendAt:     item.SendAfter,
		}
	}), nil
}

var sendWorkerMutex sync.Mutex

func SendWorker() {
	sendWorkerMutex.Lock()
	defer sendWorkerMutex.Unlock()
	// load toot ready to send
	var toots []model.Toot

	err := vars.DB.Where("send_status = ?", SENDSTATUS_WAITING).Where("send_after < ?", time.Now()).Find(&toots).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("SendWorker", err.Error())
		}
		return
	}
	for _, toot := range toots {
		id, err := sendToot(toot.TootContent, strings.Split(toot.MediaFiles, "\n"))
		if err != nil {
			toot.SendStatus = SENDSTATUS_ERROR
			toot.SendResult = err.Error()
			if err = vars.DB.Save(toot).Error; err != nil {
				fmt.Println("SendWorker", err.Error())
			}
		} else {
			toot.SendStatus = SENDSTATUS_SUCCESS
			toot.SendResult = "TOOT_ID:" + id
			if err = vars.DB.Save(toot).Error; err != nil {
				fmt.Println("SendWorker", err.Error())
			}
		}
	}
}

var cleanWorkerMutex sync.Mutex

func CleanWorker() {
	cleanWorkerMutex.Lock()
	defer cleanWorkerMutex.Unlock()
	// load send success
	var toots []model.Toot
	err := vars.DB.Find(&toots, "send_status = ?", SENDSTATUS_SUCCESS).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("CleanWorker", err.Error())
		}
		return
	}
	for _, toot := range toots {
		if err := deleteQueueItem(&toot); err != nil {
			fmt.Println("CleanWorker", err.Error())
		}
	}
}

func deleteQueueItem(toot *model.Toot) error {
	mediaFiles := strings.Split(toot.MediaFiles, "\n")
	for _, fileName := range mediaFiles {
		fileName = strings.TrimSpace(fileName)
		if len(fileName) > 0 {
			if err := os.Remove(vars.DataPath("media", fileName)); err != nil {
				return err
			}
		}
	}
	return vars.DB.Unscoped().Delete(toot).Error
}

func ClearQueue() error {
	var toots []model.Toot
	err := vars.DB.Find(&toots).Error
	if err != nil {
		return utils.IgnoreErrNotFound(err)
	}
	for _, toot := range toots {
		if err := deleteQueueItem(&toot); err != nil {
			return err
		}
	}
	return nil
}

func RemoveInQueue(id uint) error {
	var toot model.Toot
	err := vars.DB.First(&toot, id).Error
	if err != nil {
		return utils.IgnoreErrNotFound(err)
	}
	return deleteQueueItem(&toot)
}
