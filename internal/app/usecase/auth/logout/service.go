package logout

import (
	"fmt"
	"net/http"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Service(c *fiber.Ctx, userContext entity.UserContext) error
}

type service struct {
	redis redisx.Redis
}

func NewService(
	redis redisx.Redis,
) Service {
	return &service{redis: redis}
}

func (s *service) Service(c *fiber.Ctx, userContext entity.UserContext) error {
	if err := s.redis.Del(c.Context(), fmt.Sprintf("access_token:%s", userContext.User.ID.Hex())); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	if err := s.redis.Del(c.Context(), fmt.Sprintf("refresh_token:%s", userContext.User.ID.Hex())); err != nil {
		return httpx.NewErrorResponse[any](c, http.StatusInternalServerError, err)
	}

	return httpx.NewSuccessResponse[any](c, nil, int(fiber.StatusNoContent), nil)
}
