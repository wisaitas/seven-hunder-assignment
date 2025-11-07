package createuser

import (
	"errors"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Service(c *fiber.Ctx, request *Request) error
}

type service struct {
	userRepository repository.UserRepository
}

func newService(
	userRepository repository.UserRepository,
) Service {
	return &service{userRepository: userRepository}
}

func (s *service) Service(c *fiber.Ctx, request *Request) error {
	existingUser := &entity.User{}
	err := s.userRepository.FindByEmail(c, request.Email, existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	if err == nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusConflict, errors.New("email already exists"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	request.Password = string(hashedPassword)

	user := s.mapRequestToEntity(request)

	if err := s.userRepository.CreateUser(c, user); err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	return httpx.NewSuccessResponse[any](c, nil, int(fiber.StatusCreated), nil)
}
