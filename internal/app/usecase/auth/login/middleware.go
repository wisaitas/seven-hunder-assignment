package login

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
)

type Middleware struct {
	jwt       jwtx.Jwt
	redis     redisx.Redis
	validator validatorx.Validator
}

func NewMiddleware(
	jwt jwtx.Jwt,
	redis redisx.Redis,
	validator validatorx.Validator,
) *Middleware {
	return &Middleware{
		jwt:       jwt,
		redis:     redis,
		validator: validator,
	}
}

func (m *Middleware) Middleware(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	if err := m.validator.ValidateStruct(&req); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	blockKey := fmt.Sprintf("login:block:%s", req.Email)
	blocked, err := m.redis.Get(c.Context(), blockKey)
	if err == nil && blocked != "" {
		ttl, err := m.redis.TTL(c.Context(), blockKey)
		if err == nil && ttl > 0 {
			unlockTime := time.Now().Add(ttl)
			unlockTimeISO := unlockTime.Format(time.RFC3339)
			publicMsg := unlockTimeISO

			return httpx.NewErrorResponse[any](c, fiber.StatusTooManyRequests, fmt.Errorf("account locked due to too many failed login attempts"), publicMsg)
		}

		return httpx.NewErrorResponse[any](c, fiber.StatusTooManyRequests,
			fmt.Errorf("account temporarily locked due to too many failed login attempts, please try again later"))
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", req.Email)
	attemptsStr, err := m.redis.Get(c.Context(), attemptsKey)
	attempts := 0
	if err == nil && attemptsStr != "" {
		attempts, _ = strconv.Atoi(attemptsStr)
	}

	if attempts >= 3 {
		blockCountKey := fmt.Sprintf("login:block_count:%s", req.Email)
		blockCountStr, _ := m.redis.Get(c.Context(), blockCountKey)
		blockCount := 0
		if blockCountStr != "" {
			blockCount, _ = strconv.Atoi(blockCountStr)
		}

		blockDuration := time.Duration(5*math.Pow(2, float64(blockCount))) * time.Minute

		maxDuration := 24 * time.Hour
		if blockDuration > maxDuration {
			blockDuration = maxDuration
		}

		if err := m.redis.Set(c.Context(), blockKey, "1", blockDuration); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
		}

		blockCount++
		if err := m.redis.Set(c.Context(), blockCountKey, strconv.Itoa(blockCount), 7*24*time.Hour); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
		}

		if err := m.redis.Del(c.Context(), attemptsKey); err != nil {
			return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
		}

		unlockTime := time.Now().Add(blockDuration)
		unlockTimeISO := unlockTime.Format(time.RFC3339)
		publicMsg := unlockTimeISO

		return httpx.NewErrorResponse[any](c, fiber.StatusTooManyRequests, fmt.Errorf("account locked due to too many failed login attempts"), publicMsg)
	}

	c.Locals("req", req)
	return c.Next()
}
