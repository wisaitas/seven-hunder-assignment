package auth

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/register"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
)

type UseCase struct {
	Register *register.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		Register: register.New(userRepository, validator),
	}
}
