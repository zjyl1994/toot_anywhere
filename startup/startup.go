package startup

import (
	"encoding/hex"
	"errors"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/robfig/cron/v3"
	"github.com/zjyl1994/toot_anywhere/model"
	"github.com/zjyl1994/toot_anywhere/server"
	"github.com/zjyl1994/toot_anywhere/service/config"
	"github.com/zjyl1994/toot_anywhere/service/toot"
	"github.com/zjyl1994/toot_anywhere/vars"
	"github.com/zjyl1994/utilz"
	"gorm.io/gorm"
)

func Main() (err error) {
	port := utilz.GetEnvString("TOOT_PORT")
	listenAddr := utilz.GetEnvString("TOOT_PORT") + ":" +
		utilz.If(len(port) > 0, port, vars.DEFAULT_PORT)

	err = os.MkdirAll(vars.DataPath(), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(vars.DataPath(toot.DATA_DIR), 0755)
	if err != nil {
		return err
	}

	vars.DB, err = gorm.Open(sqlite.Open(vars.DataPath("database.sqlite3")), &gorm.Config{})
	if err != nil {
		return err
	}

	err = vars.DB.AutoMigrate(&model.Toot{}, &model.Config{})
	if err != nil {
		return err
	}
	// init secret
	secret, err := config.Get("secret")
	if err == nil {
		vars.Secret = secret
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			vars.Secret = hex.EncodeToString(utilz.RandBytes(32))
			if err := config.Set("secret", vars.Secret); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// init cron worker
	vars.Cron = cron.New()
	vars.Cron.AddFunc("@every 1m", toot.SendWorker)
	vars.Cron.AddFunc("@every 1h", toot.CleanWorker)
	vars.Cron.Start()

	return server.Run(listenAddr)
}
