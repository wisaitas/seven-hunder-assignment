package updateuser

import (
	"errors"

	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	param := c.Params("user_id")
	if param == "" {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, errors.New("user ID is required"))
	}

	userID, err := bson.ObjectIDFromHex(param)
	if err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	request := Request{}
	if err := c.BodyParser(&request); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	if err := h.validator.ValidateStruct(&request); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	return h.service.Service(c, userID, &request)
}
