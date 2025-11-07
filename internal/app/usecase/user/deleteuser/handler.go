package deleteuser

import (
	"errors"

	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	param := c.Params("user_id")
	if param == "" {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, errors.New("user ID is required"))
	}

	userID, err := bson.ObjectIDFromHex(param)
	if err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, err)
	}

	return h.service.Service(c, userID)
}
