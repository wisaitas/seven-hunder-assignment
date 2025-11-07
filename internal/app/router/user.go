package router

import (
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user"
	"github.com/gofiber/fiber/v2"
)

type UserRouter struct {
	apiRouter   fiber.Router
	userUseCase *user.UseCase
}

func NewUserRouter(
	apiRouter fiber.Router,
	userUseCase *user.UseCase,
) *UserRouter {
	return &UserRouter{
		apiRouter:   apiRouter,
		userUseCase: userUseCase,
	}
}

func (r *UserRouter) Setup() {
	userRouter := r.apiRouter.Group("/users")

	userRouter.Get("/", r.userUseCase.GetUsersMiddleware, r.userUseCase.GetUsers.Handle)
	userRouter.Get("/:user_id", r.userUseCase.GetUserByIDMiddleware, r.userUseCase.GetUserByID.Handle)

	userRouter.Post("/", r.userUseCase.CreateUserMiddleware, r.userUseCase.CreateUser.Handle)

	userRouter.Patch("/:user_id", r.userUseCase.UpdateUserMiddleware, r.userUseCase.UpdateUser.Handle)

	userRouter.Delete("/:user_id", r.userUseCase.DeleteUserMiddleware, r.userUseCase.DeleteUser.Handle)
}
