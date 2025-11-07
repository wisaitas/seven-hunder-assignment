// internal/app/usecase/auth/use_case.go
package auth

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/login"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/logout"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/register"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
)

type UseCase struct {
	Register *register.Handler
	Login    *login.Handler
	Logout   *logout.Handler

	LoginMiddleware  fiber.Handler
	LogoutMiddleware fiber.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
	jwt jwtx.Jwt,
	redis redisx.Redis,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		Register: register.NewHandler(userRepository, validator),

		Login:           login.NewHandler(userRepository, jwt, redis, validator),
		LoginMiddleware: login.NewMiddleware(jwt, redis, validator),

		Logout:           logout.NewHandler(redis),
		LogoutMiddleware: logout.NewMiddleware(jwt, redis),
	}
}
