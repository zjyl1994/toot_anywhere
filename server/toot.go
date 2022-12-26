package server

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"github.com/zjyl1994/toot_anywhere/service/toot"
	"github.com/zjyl1994/toot_anywhere/vars"
	"github.com/zjyl1994/utilz"
)

func TootPage(c *fiber.Ctx) error {
	return c.Render("toot", fiber.Map{"PageTitle": "嘟嘟"})
}

func TootHandler(c *fiber.Ctx) error {
	content := c.FormValue("toot")
	delaySec, err := strconv.Atoi(c.FormValue("delay"))
	if err != nil {
		delaySec = 0
	}
	mf, err := c.MultipartForm()
	if err != nil {
		return err
	}
	var mediaFilePath []string
	if medias, ok := mf.File["media"]; ok {
		for _, media := range medias {
			err := c.SaveFile(media, vars.DataPath(toot.DATA_DIR, media.Filename))
			if err != nil {
				return err
			}
			mediaFilePath = append(mediaFilePath, media.Filename)
		}
	}
	return toot.PushQueue(content, mediaFilePath, delaySec)
}

type queueRenderItem struct {
	ID         uint     `json:"id"`
	Content    string   `json:"content"`
	MediaFiles []string `json:"media"`
	SendStatus string   `json:"status"`
	SendResult string   `json:"result"`
}

func TootQueuePage(c *fiber.Ctx) error {
	return c.Render("queue", fiber.Map{"PageTitle": "发送队列"})
}

func TootQueueHandler(c *fiber.Ctx) error {
	items, err := toot.ListQueue()
	if err != nil {
		return err
	}
	renderItems := lo.Map(items, func(x toot.QueueItem, _ int) queueRenderItem {
		var status string
		switch x.SendStatus {
		case toot.SENDSTATUS_WAITING:
			status = fmt.Sprintf("等待发送 (计划: %s)", utilz.FormatTime(x.SendAt))
		case toot.SENDSTATUS_SUCCESS:
			status = "发送成功"
		case toot.SENDSTATUS_ERROR:
			status = "发送失败"
		default:
			status = "未知"
		}
		return queueRenderItem{
			ID:         x.ID,
			Content:    x.Content,
			MediaFiles: lo.Map(x.Media, func(x string, _ int) string { return "/media/" + x }),
			SendStatus: status,
			SendResult: x.SendResult,
		}
	})
	return c.JSON(renderItems)
}

func TootRemoveHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	queueId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return toot.RemoveInQueue(uint(queueId))
}
