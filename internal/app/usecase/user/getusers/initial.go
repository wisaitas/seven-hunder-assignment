package getusers

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
)

func New(
	userRepository repository.UserRepository,
) *Handler {
	service := NewService(userRepository)

	return NewHandler(service)
}
