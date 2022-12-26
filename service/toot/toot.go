package toot

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/zjyl1994/toot_anywhere/service/config"
	"github.com/zjyl1994/toot_anywhere/service/toot/mastodon"
	"github.com/zjyl1994/toot_anywhere/vars"
	"gorm.io/gorm"
)

type sendTootHandlerFunc func(content string, media []string) (string, error)

var sendTootHandlerMap = map[string]sendTootHandlerFunc{
	"mastodon": mastodon.SendToot,
}

const (
	CONFIG_SEND_HANDLER = "sendhandler"
)

func sendToot(content string, media []string) (string, error) {
	handlerName, err := config.Get(CONFIG_SEND_HANDLER)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("toot handler not setup")
		} else {
			return "", err
		}
	}
	handlerFn, ok := sendTootHandlerMap[handlerName]
	if !ok {
		return "", fmt.Errorf(`toot handler "%s" not found`, handlerName)
	}
	mediaFiles := lo.Map(media, func(x string, _ int) string {
		return vars.DataPath(DATA_DIR, x)
	})
	return handlerFn(content, mediaFiles)
}
