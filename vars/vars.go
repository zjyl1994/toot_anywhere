package vars

import (
	"path/filepath"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	APP_NAME        = "TootAnywhere"
	DEFAULT_PORT    = "9900"
	DATA_VOLUME_DIR = "data"
)

var (
	DB      *gorm.DB
	Cron    *cron.Cron
	Secret  string
	Setuped bool
)

func DataPath(path ...string) string {
	return filepath.Join(append([]string{DATA_VOLUME_DIR}, path...)...)
}
