package initial

import (
	appConfig "github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/middleware"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
)

func newMiddleware(app *fiber.App, client *client) {
	app.Use(middleware.NewCors())
	app.Use(httpx.NewLogger(appConfig.Config.Service.Name, httpx.WithMaskMap(appConfig.Config.Service.MaskMap)))
	app.Use(middleware.NewRateLimit())
	app.Use(middleware.Healthz(client.mongoDB, client.redis.Client()))
	app.Use(middleware.Recover())
}
