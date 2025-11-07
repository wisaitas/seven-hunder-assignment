package updateuser

import (
	"errors"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service interface {
	Service(c *fiber.Ctx, userID bson.ObjectID, request *Request) error
}

type service struct {
	userRepository repository.UserRepository
}

func NewService(
	userRepository repository.UserRepository,
) Service {
	return &service{userRepository: userRepository}
}

func (s *service) Service(c *fiber.Ctx, userID bson.ObjectID, request *Request) error {
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

	updates := bson.M{}

	if request.Name != nil {
		updates["name"] = *request.Name
	}

	if request.Email != nil {
		existingUser := &entity.User{}
		err := s.userRepository.FindByEmail(c, *request.Email, existingUser)

		if err != nil && err != mongo.ErrNoDocuments {
			return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
		}

		if err == nil && existingUser.ID != userID {
			return httpx.NewErrorResponse[any](c, fiber.StatusConflict, errors.New("email already exists"))
		}

		updates["email"] = *request.Email
	}

	if len(updates) == 0 {
		return httpx.NewErrorResponse[any](c, fiber.StatusBadRequest, errors.New("no fields to update"))
	}

	updates["updated_at"] = bson.DateTime(time.Now().UnixMilli())

	if err := s.userRepository.UpdateUser(c, userID, user.Version, updates); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	return httpx.NewSuccessResponse[any](c, nil, int(fiber.StatusNoContent), nil)
}
