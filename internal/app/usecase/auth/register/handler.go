package register

import (
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service   Service
	validator validatorx.Validator
}

func NewHandler(
	service Service,
	validator validatorx.Validator,
) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	request := &Request{}
	if err := c.BodyParser(request); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	if err := h.validator.ValidateStruct(request); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	return h.service.Service(c, request)
}
