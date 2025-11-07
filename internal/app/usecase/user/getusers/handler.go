package getusers

import (
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(
	service Service,
) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	queryParam := QueryParam{}
	if err := c.QueryParser(&queryParam); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	if err := httpx.SetDefaultPagination(&queryParam); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	return h.service.Service(c, queryParam)
}
