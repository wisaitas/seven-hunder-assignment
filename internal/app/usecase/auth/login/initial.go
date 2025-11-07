package login

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
)

func NewHandler(
	userRepository repository.UserRepository,
	jwt jwtx.Jwt,
	redis redisx.Redis,
	validator validatorx.Validator,
) *Handler {
	service := NewService(userRepository)
	handler := newHandler(service, validator)

	return handler
}
