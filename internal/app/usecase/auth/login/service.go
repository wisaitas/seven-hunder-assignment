package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Service(c *fiber.Ctx, request *Request) error
}

type service struct {
	userRepository repository.UserRepository
	redis          redisx.Redis
	jwt            jwtx.Jwt
}

func NewService(
	userRepository repository.UserRepository,
	redis redisx.Redis,
	jwt jwtx.Jwt,
) Service {
	return &service{
		userRepository: userRepository,
		redis:          redis,
		jwt:            jwt,
	}
}

func (s *service) Service(c *fiber.Ctx, request *Request) error {
	attemptsKey := fmt.Sprintf("login:attempts:%s", request.Email)

	user := &entity.User{}
	if err := s.userRepository.FindByEmail(c, request.Email, user); err != nil {
		if err := s.incrementFailedAttempts(c.Context(), attemptsKey); err != nil {
			return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
		}

		if err == mongo.ErrNoDocuments {
			return httpx.NewErrorResponse[any](c, fiber.StatusNotFound, errors.New("user not found"))
		}
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		if err := s.incrementFailedAttempts(c.Context(), attemptsKey); err != nil {
			return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
		}

		return httpx.NewErrorResponse[any](c, int(fiber.StatusUnauthorized), errors.New("password is incorrect"))
	}

	if err := s.redis.Del(c.Context(), attemptsKey); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	if err := s.redis.Del(c.Context(), fmt.Sprintf("login:block_count:%s", request.Email)); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	timeNow := time.Now()
	accessTokenExpired := timeNow.Add(time.Duration(app.Config.JWT.AccessTTL) * time.Minute)
	refreshTokenExpired := timeNow.Add(time.Duration(app.Config.JWT.RefreshTTL) * time.Hour)

	externalContext := entity.ExternalContext{
		Subject:   user.ID.Hex(),
		ExpiresAt: accessTokenExpired,
		IssuedAt:  timeNow,
	}

	accessToken, err := s.jwt.Generate(externalContext, app.Config.JWT.Secret)
	if err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	externalContext.ExpiresAt = refreshTokenExpired
	refreshToken, err := s.jwt.Generate(externalContext, app.Config.JWT.Secret)
	if err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	userContext := entity.UserContext{
		User: *user,
	}

	userContextJSON, err := json.Marshal(userContext)
	if err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	if err := s.redis.Set(context.Background(), fmt.Sprintf("access_token:%s", user.ID.Hex()), string(userContextJSON), accessTokenExpired.Sub(timeNow)); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	if err := s.redis.Set(context.Background(), fmt.Sprintf("refresh_token:%s", user.ID.Hex()), string(userContextJSON), refreshTokenExpired.Sub(timeNow)); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	return httpx.NewSuccessResponse(c, &Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, int(fiber.StatusOK), nil)
}

func (s *service) incrementFailedAttempts(ctx context.Context, attemptsKey string) error {
	attemptsStr, _ := s.redis.Get(ctx, attemptsKey)
	attempts := 0
	if attemptsStr != "" {
		attempts, _ = strconv.Atoi(attemptsStr)
	}
	attempts++

	if err := s.redis.Set(ctx, attemptsKey, strconv.Itoa(attempts), 15*time.Minute); err != nil {
		return err
	}

	return nil
}
