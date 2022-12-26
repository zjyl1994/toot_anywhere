package server

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zjyl1994/toot_anywhere/service/config"
	"github.com/zjyl1994/toot_anywhere/service/toot"
	"github.com/zjyl1994/toot_anywhere/vars"
	"github.com/zjyl1994/utilz"
)

func SettingPage(c *fiber.Ctx) error {
	mastodonUrl, _ := config.Get("mastodon.url")
	mastodonToken, _ := config.Get("mastodon.access_token")
	return c.Render("setting", fiber.Map{
		"PageTitle":   "设置面板",
		"Instance":    mastodonUrl,
		"AccessToken": mastodonToken,
	})
}

func SettingPasswordHandler(c *fiber.Ctx) error {
	password := c.FormValue("password")
	repeated := c.FormValue("repeated")
	if password != repeated {
		return fiber.NewError(http.StatusBadRequest, "两次密码不一致")
	}
	hash, err := utilz.PasswordHash(password)
	if err != nil {
		return err
	}
	err = config.Set("password", hash)
	if err != nil {
		return err
	}
	return c.Render("message", fiber.Map{
		"PageTitle": "消息",
		"Title":     "成功",
		"Color":     "success",
		"Content":   "密码成功更新",
	})
}

func SettingMastodonHandler(c *fiber.Ctx) error {
	instance := c.FormValue("instance")
	accesstoken := c.FormValue("accesstoken")
	if instance == "" {
		return fiber.NewError(http.StatusBadRequest, "实例网址不能为空")
	}
	if accesstoken == "" {
		return fiber.NewError(http.StatusBadRequest, "AccessToken不能为空")
	}
	instance = strings.ToLower(instance)

	if err := config.Set("mastodon.url", instance); err != nil {
		return err
	}
	if err := config.Set("mastodon.access_token", accesstoken); err != nil {
		return err
	}
	if err := config.Set("sendhandler", "mastodon"); err != nil {
		return err
	}

	return c.Render("message", fiber.Map{
		"PageTitle": "消息",
		"Title":     "成功",
		"Color":     "success",
		"Content":   "Mastodon 设置成功更新",
	})
}

// ============

func SetupMiddleware(c *fiber.Ctx) error {
	if vars.Setuped {
		return c.Next()
	}
	isset, err := config.IsSet(toot.CONFIG_SEND_HANDLER)
	if err != nil {
		return err
	}
	if isset {
		vars.Setuped = true
		return c.Next()
	}
	return c.Redirect("/setup")
}

func SetupPage(c *fiber.Ctx) error {
	if vars.Setuped {
		return c.Redirect("/")
	}
	return c.Render("setup", nil, "")
}

func SetupHandler(c *fiber.Ctx) error {
	if vars.Setuped {
		return c.Redirect("/")
	}
	pass := c.FormValue("password")
	instanceUrl := c.FormValue("instance")
	instanceToken := c.FormValue("accesstoken")

	hash, err := utilz.PasswordHash(pass)
	if err != nil {
		return err
	}
	err = config.Set("password", hash)
	if err != nil {
		return err
	}
	if err := config.Set("mastodon.url", instanceUrl); err != nil {
		return err
	}
	if err := config.Set("mastodon.access_token", instanceToken); err != nil {
		return err
	}
	if err := config.Set(toot.CONFIG_SEND_HANDLER, "mastodon"); err != nil {
		return err
	}
	setAuthCookie(c, false)
	vars.Setuped = true
	return c.Redirect("/")
}
