package logout

import "github.com/7-solutions/backend-challenge/pkg/db/redisx"

func NewHandler(
	redis redisx.Redis,
) *Handler {
	service := NewService(redis)
	handler := newHandler(service)

	return handler
}
