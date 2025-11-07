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
	jwtLib "github.com/golang-jwt/jwt/v5"
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
	_, err := jwt.Parse(token, &tokenContext, secret)
	if err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	userContextJSON, err := redis.Get(context.Background(), fmt.Sprintf("access_token:%s", tokenContext.ID))
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
	_, err := jwt.Parse(token, &tokenContext, secret)
	if err != nil {
		return fmt.Errorf("[util-auth] %w", err)
	}

	userContextJSON, err := redis.Get(context.Background(), fmt.Sprintf("refresh_token:%s", tokenContext.ID))
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

func GenerateToken(data map[string]interface{}, exp int64, secret string) (string, error) {
	claim := jwtLib.MapClaims(data)
	claim["exp"] = exp
	claim["iat"] = time.Now().Unix()

	token := jwtLib.NewWithClaims(jwtLib.SigningMethodHS256, claim)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("[util-auth] %w", err)
	}

	return tokenString, nil
}
