package router

import (
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth"
	"github.com/gofiber/fiber/v2"
)

type AuthRouter struct {
	apiRouter   fiber.Router
	authUseCase *auth.UseCase
}

func NewAuthRouter(
	apiRouter fiber.Router,
	authUseCase *auth.UseCase,
) *AuthRouter {
	return &AuthRouter{
		apiRouter:   apiRouter,
		authUseCase: authUseCase,
	}
}

func (r *AuthRouter) Setup() {
	userRouter := r.apiRouter.Group("/auth")
	userRouter.Post("/register", r.authUseCase.Register.Handle)
}
