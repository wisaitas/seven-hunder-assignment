package login

import (
	"errors"

	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Service(c *fiber.Ctx, request *Request) error
}

type service struct {
	userRepository repository.UserRepository
}

func NewService(
	userRepository repository.UserRepository,
) Service {
	return &service{
		userRepository: userRepository,
	}
}

func (s *service) Service(c *fiber.Ctx, request *Request) error {
	user, err := s.userRepository.FindByEmail(c, request.Email)
	if err != nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusInternalServerError, err)
	}

	if user == nil {
		return httpx.NewErrorResponse[any](c, fiber.StatusNotFound, errors.New("user not found"))
	}

	return httpx.NewSuccessResponse[any](c, nil, int(fiber.StatusOK), nil)
}
