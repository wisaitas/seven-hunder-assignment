package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Healthz(
	mongoDB *mongo.Client,
	redis *redis.Client,
) fiber.Handler {
	return healthcheck.New(
		healthcheck.Config{
			LivenessEndpoint: "/api/v1/livez",
			LivenessProbe: func(c *fiber.Ctx) bool {
				return true
			},
			ReadinessEndpoint: "/api/v1/readyz",
			ReadinessProbe: func(c *fiber.Ctx) bool {
				if err := mongoDB.Ping(c.Context(), nil); err != nil {
					fmt.Println("[readiness probe] mongoDB ping error", err)
					return false
				}
				if err := redis.Ping(c.Context()).Err(); err != nil {
					fmt.Println("[readiness probe] redis ping error", err)
					return false
				}
				return true
			},
		},
	)
}
