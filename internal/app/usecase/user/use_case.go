package user

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/deleteuser"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getuserbyid"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getusers"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/updateuser"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"github.com/gofiber/fiber/v2"
)

type UseCase struct {
	GetUsers           *getusers.Handler
	GetUsersMiddleware fiber.Handler

	GetUserByID           *getuserbyid.Handler
	GetUserByIDMiddleware fiber.Handler

	UpdateUser           *updateuser.Handler
	UpdateUserMiddleware fiber.Handler

	DeleteUser           *deleteuser.Handler
	DeleteUserMiddleware fiber.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
	jwt jwtx.Jwt,
	redis redisx.Redis,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		GetUsers:           getusers.New(userRepository),
		GetUsersMiddleware: getusers.NewMiddleware(jwt, redis),

		GetUserByID:           getuserbyid.New(userRepository),
		GetUserByIDMiddleware: getuserbyid.NewMiddleware(jwt, redis),

		UpdateUser:           updateuser.New(userRepository, validator),
		UpdateUserMiddleware: updateuser.NewMiddleware(jwt, redis),

		DeleteUser:           deleteuser.New(userRepository),
		DeleteUserMiddleware: deleteuser.NewMiddleware(jwt, redis),
	}
}
