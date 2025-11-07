package util

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	redisLib "github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v2"
)

func AuthAccessToken(c *fiber.Ctx, redis redisx.Redis, jwt jwtx.Jwt, secret string) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("[util-auth] invalid token type")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	var tokenContext entity.ExternalContext
	err := jwt.Parse(token, secret, &tokenContext)
	if err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	// ตรวจสอบว่า token หมดอายุหรือไม่
	if time.Now().After(tokenContext.ExpiresAt) {
		return fmt.Errorf("[util-auth] token expired")
	}

	userContextJSON, err := redis.Get(context.Background(), fmt.Sprintf("access_token:%s", tokenContext.Subject))
	if err != nil {
		if err == redisLib.Nil {
			return fmt.Errorf("[util-auth] session not found")
		}

		return fmt.Errorf("[util-auth] %w", err)
	}

	var userContext entity.UserContext
	if err := json.Unmarshal([]byte(userContextJSON), &userContext); err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	c.Locals("userContext", userContext)
	return nil
}

func AuthRefreshToken(c *fiber.Ctx, redis redisx.Redis, jwt jwtx.Jwt, secret string) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return fmt.Errorf("[util-auth] invalid token type")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	var tokenContext entity.ExternalContext
	err := jwt.Parse(token, secret, &tokenContext)
	if err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	// ตรวจสอบว่า token หมดอายุหรือไม่
	if time.Now().After(tokenContext.ExpiresAt) {
		return fmt.Errorf("[util-auth] token expired")
	}

	userContextJSON, err := redis.Get(context.Background(), fmt.Sprintf("refresh_token:%s", tokenContext.Subject))
	if err != nil {
		if err == redisLib.Nil {
			return fmt.Errorf("[util-auth] session not found")
		}

		return fmt.Errorf("[util-auth] %w", err)
	}

	var userContext entity.UserContext
	if err := json.Unmarshal([]byte(userContextJSON), &userContext); err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	c.Locals("userContext", userContext)
	return nil
}

func GenerateToken(data entity.ExternalContext, jwt jwtx.Jwt, secret string) (string, error) {
	tokenString, err := jwt.Generate(data, secret)
	if err != nil {
		return "", fmt.Errorf("[util-auth] %w", err)
	}

	return tokenString, nil
}
