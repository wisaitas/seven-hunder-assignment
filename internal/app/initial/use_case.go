package initial

import (
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user"
)

type useCase struct {
	authUseCase *auth.UseCase
	userUseCase *user.UseCase
}

func newUseCase(
	repository *repository,
	sdk *sdk,
) *useCase {
	return &useCase{
		authUseCase: auth.NewUseCase(repository.userRepository, sdk.validator),
		userUseCase: user.NewUseCase(repository.userRepository),
	}
}
