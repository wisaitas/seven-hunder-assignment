package login

import (
	"errors"

	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service   Service
	validator validatorx.Validator
}

func newHandler(
	service Service,
	validator validatorx.Validator,
) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	req, ok := c.Locals("req").(Request)
	if !ok {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, errors.New("request not found"))
	}

	return h.service.Service(c, &req)
}
