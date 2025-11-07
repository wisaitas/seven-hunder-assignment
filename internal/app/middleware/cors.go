package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewCors() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowHeaders:  "Authorization, Content-Type, Accept, Origin, X-Requested-With, Range",
		AllowMethods:  "GET, POST, PUT, DELETE, OPTIONS",
		ExposeHeaders: "Content-Disposition, Content-Length",
	})
}
