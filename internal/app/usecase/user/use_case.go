package user

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getusers"
)

type UseCase struct {
	GetUsers *getusers.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
) *UseCase {
	return &UseCase{
		GetUsers: getusers.New(userRepository),
	}
}
