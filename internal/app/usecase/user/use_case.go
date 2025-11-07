package user

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/deleteuser"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getuserbyid"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getusers"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/updateuser"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
)

type UseCase struct {
	GetUsers    *getusers.Handler
	GetUserByID *getuserbyid.Handler

	UpdateUser *updateuser.Handler

	DeleteUser *deleteuser.Handler
}

func NewUseCase(
	userRepository repository.UserRepository,
	validator validatorx.Validator,
) *UseCase {
	return &UseCase{
		GetUsers:    getusers.New(userRepository),
		GetUserByID: getuserbyid.New(userRepository),
		UpdateUser:  updateuser.New(userRepository, validator),
		DeleteUser:  deleteuser.New(userRepository),
	}
}
