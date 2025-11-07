package auth

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/login"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/register"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
)

type UseCase struct {
	Login           *login.Handler
	LoginMiddleware *login.Middleware
	Register        *register.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
	jwt jwtx.Jwt,
	redis redisx.Redis,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		Login:           login.NewHandler(userRepository, jwt, redis, validator),
		LoginMiddleware: login.NewMiddleware(jwt, redis, validator),
		Register:        register.NewHandler(userRepository, validator),
	}
}
