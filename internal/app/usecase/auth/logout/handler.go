package logout

import (
	"errors"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func newHandler(
	service Service,
) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	userContext, ok := c.Locals("userContext").(entity.UserContext)
	if !ok {
		return httpx.NewErrorResponse[any](c, fiber.StatusUnauthorized, errors.New("user context not found"))
	}

	return h.service.Service(c, userContext)
}
