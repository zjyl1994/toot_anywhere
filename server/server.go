package server

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	"github.com/gofiber/utils"
	"github.com/zjyl1994/toot_anywhere/service/toot"
	"github.com/zjyl1994/toot_anywhere/vars"
)

func Run(listen string) error {
	app := fiber.New(fiber.Config{
		Views:        html.New("./web/tmpl", ".html").Reload(true),
		ViewsLayout:  "layout",
		ErrorHandler: defaultErrorHandler,
	})
	app.Use(recover.New())
	app.Use(favicon.New(favicon.Config{
		File: "./web/static/favicon.ico",
	}))
	app.Static("/media", vars.DataPath(toot.DATA_DIR))

	app.Get("/setup", SetupPage)
	app.Post("/setup", SetupHandler)

	app.Use(SetupMiddleware)

	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/toot") })

	app.Get("/login", LoginPage)
	app.Post("/login", LoginHandler)
	app.Get("/logout", LogoutHandler)

	app.Use(AuthMiddleware)

	app.Get("/setting", SettingPage)
	app.Post("/setting/password", SettingPasswordHandler)
	app.Post("/setting/mastodon", SettingMastodonHandler)

	app.Get("/toot", TootPage)
	app.Post("/toot", TootHandler)

	app.Get("/queue", TootQueuePage)
	app.Get("/queue/list", TootQueueHandler)
	app.Delete("/queue/:id", TootRemoveHandler)

	return app.Listen(listen)
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).Render("message",
		fiber.Map{
			"PageTitle": strconv.Itoa(code),
			"Title":     utils.StatusMessage(code),
			"Content":   err.Error(),
			"Color":     "danger",
		},
	)
}
