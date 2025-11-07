package getuserbyid

import (
	"errors"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service interface {
	Service(c *fiber.Ctx, userID bson.ObjectID) error
}

type service struct {
	userRepository repository.UserRepository
}

func NewService(
	userRepository repository.UserRepository,
) Service {
	return &service{userRepository: userRepository}
}

func (s *service) Service(c *fiber.Ctx, userID bson.ObjectID) error {
	var user entity.User

	filter := bson.M{
		"_id":        userID,
		"deleted_at": nil,
	}

	if err := s.userRepository.FindByID(c, filter, &user); err != nil {
		if err == mongo.ErrNoDocuments {
			return httpx.NewErrorResponse[any](c, fiber.StatusNotFound, errors.New("user not found"))
		}

		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	resp := s.mapEntityToResponse(&user)

	return httpx.NewSuccessResponse(c, &resp, int(fiber.StatusOK), nil)
}
