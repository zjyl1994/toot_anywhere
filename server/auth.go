package server

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zjyl1994/toot_anywhere/service/config"
	"github.com/zjyl1994/toot_anywhere/vars"
	"github.com/zjyl1994/utilz"
)

const COOKIE_NAME_LOGIN_FLAG = "token"

func LoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{"PageTitle": "登录"}, "")
}

func LoginHandler(c *fiber.Ctx) error {
	passHash, err := config.Get("password")
	if err != nil {
		return err
	}

	password := c.FormValue("password")
	if !utilz.CheckPasswordHash(password, passHash) {
		return fiber.NewError(http.StatusForbidden, "密码错误")
	}

	setAuthCookie(c, c.FormValue("keeplogin") == "remember")
	return c.Redirect("/")
}

func LogoutHandler(c *fiber.Ctx) error {
	c.ClearCookie()
	return c.Redirect("/login")
}

func AuthMiddleware(c *fiber.Ctx) error {
	if isPasswordSet, err := config.IsSet("password"); err != nil {
		return err
	} else if !isPasswordSet { // Skip auth when password not set
		return c.Next()
	}

	loginToken := c.Cookies(COOKIE_NAME_LOGIN_FLAG)
	var token jsonToken
	err := utilz.FromJSONString(loginToken, &token)
	if err != nil {
		return c.Redirect("/login")
	}
	if token.Signature != token.Sign(vars.Secret) {
		return c.Redirect("/login")
	}
	return c.Next()
}

type jsonToken struct {
	Nonce     string
	Expire    time.Time
	Signature string
}

func (p *jsonToken) Sign(secret string) string {
	h := sha256.New()
	io.WriteString(h, p.Nonce)
	io.WriteString(h, "$")
	io.WriteString(h, utilz.FormatTime(p.Expire))
	io.WriteString(h, "$")
	io.WriteString(h, secret)
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}

func setAuthCookie(c *fiber.Ctx, rememberMe bool) {
	tokenExpireAt := time.Now().AddDate(0, 1, 0)
	token := jsonToken{
		Nonce:  utilz.RandString(16),
		Expire: tokenExpireAt,
	}
	token.Signature = token.Sign(vars.Secret)

	cookie := new(fiber.Cookie)
	cookie.Name = COOKIE_NAME_LOGIN_FLAG
	cookie.Value = utilz.ToJSONStringNoError(token)
	if rememberMe {
		cookie.Expires = tokenExpireAt
	}
	cookie.HTTPOnly = true
	c.Cookie(cookie)
}
