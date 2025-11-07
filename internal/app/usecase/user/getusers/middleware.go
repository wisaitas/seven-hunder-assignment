package getusers

import (
	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/util"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/gofiber/fiber/v2"
)

func NewMiddleware(
	jwt jwtx.Jwt,
	redis redisx.Redis,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := util.AuthAccessToken(c, redis, jwt, app.Config.JWT.Secret); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusUnauthorized, err)
		}

		return c.Next()
	}
}
