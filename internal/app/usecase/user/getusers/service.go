// internal/app/usecase/user/getusers/service.go
package getusers

import (
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	Service(c *fiber.Ctx, queryParam QueryParam) error
}

type service struct {
	userRepository repository.UserRepository
}

func NewService(
	userRepository repository.UserRepository,
) Service {
	return &service{userRepository: userRepository}
}

func (s *service) Service(c *fiber.Ctx, queryParam QueryParam) error {
	filter := bson.M{
		"deleted_at": nil,
	}

	sortField := "created_at"
	sortOrder := -1

	users, hasNext, hasPrev, err := s.userRepository.FindAllPaginated(
		c,
		filter,
		sortField,
		sortOrder,
		*queryParam.Page,
		*queryParam.PageSize,
	)
	if err != nil {
		return httpx.NewErrorResponse[any](c, int(fiber.StatusInternalServerError), err)
	}

	actualUsers := users
	if hasNext {
		actualUsers = users[:*queryParam.PageSize]
	}

	windowWidth := 5
	half := windowWidth / 2

	wantRight := half
	if !hasPrev {
		wantRight = windowWidth - 1
	}

	nextPagesAvail, err := httpx.ProbeNextPagesMongo(
		c.Context(),
		s.userRepository.GetCollection(),
		filter,
		sortField,
		sortOrder,
		*queryParam.Page,
		*queryParam.PageSize,
		wantRight,
	)
	if err != nil {
		return httpx.NewErrorResponse[any](c, int(fiber.StatusInternalServerError), err)
	}

	resp := s.mapEntityToResponse(actualUsers)

	return httpx.NewSuccessResponse(c, &resp, int(fiber.StatusOK), &httpx.Pagination{
		TotalElements: len(actualUsers),
		Page:          *queryParam.Page,
		PageSize:      *queryParam.PageSize,
		HasPrev:       hasPrev,
		HasNext:       hasNext,
		Windows:       httpx.PageWindowClamped(*queryParam.Page, windowWidth, hasPrev, hasNext, nextPagesAvail),
	})
}
