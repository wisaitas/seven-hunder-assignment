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
	client *client,
	repository *repository,
	sdk *sdk,
) *useCase {
	return &useCase{
		authUseCase: auth.NewUseCase(repository.userRepository, sdk.jwt, client.redis, sdk.validator),
		userUseCase: user.NewUseCase(repository.userRepository, sdk.validator),
	}
}
