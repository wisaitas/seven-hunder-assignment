package updateuser

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
)

func New(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *Handler {
	service := NewService(userRepository)
	handler := newHandler(service, validator)

	return handler
}
