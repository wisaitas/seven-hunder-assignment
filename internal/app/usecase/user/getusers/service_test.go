package getusers_test

import (
	"errors"
	"testing"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getusers"
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type getUsersTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	service      getusers.Service
	app          *fiber.App
}

func (s *getUsersTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.service = getusers.NewService(s.mockUserRepo)
	s.app = fiber.New()
}

func (s *getUsersTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *getUsersTestSuite) TestGetUsersDatabaseError() {
	page := 1
	pageSize := 10
	queryParam := getusers.QueryParam{
		PaginationQuery: httpx.PaginationQuery{
			Page:     &page,
			PageSize: &pageSize,
		},
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindAllPaginated",
		ctx,
		mock.AnythingOfType("bson.M"),
		"created_at",
		-1,
		page,
		pageSize,
	).Return([]*entity.User{}, false, false, errors.New("database connection error"))

	err := s.service.Service(ctx, queryParam)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *getUsersTestSuite) TestGetUsersFindAllPaginatedError() {
	page := 2
	pageSize := 5
	queryParam := getusers.QueryParam{
		PaginationQuery: httpx.PaginationQuery{
			Page:     &page,
			PageSize: &pageSize,
		},
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindAllPaginated",
		ctx,
		mock.MatchedBy(func(filter interface{}) bool {
			return true
		}),
		"created_at",
		-1,
		page,
		pageSize,
	).Return([]*entity.User{}, false, false, errors.New("query execution error"))

	err := s.service.Service(ctx, queryParam)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *getUsersTestSuite) TestGetUsersWithDifferentPageSizes() {
	testCases := []struct {
		name     string
		page     int
		pageSize int
	}{
		{
			name:     "Page 1 Size 10",
			page:     1,
			pageSize: 10,
		},
		{
			name:     "Page 3 Size 20",
			page:     3,
			pageSize: 20,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			queryParam := getusers.QueryParam{
				PaginationQuery: httpx.PaginationQuery{
					Page:     &tc.page,
					PageSize: &tc.pageSize,
				},
			}

			ctx := s.createTestContext()
			defer s.app.ReleaseCtx(ctx)

			s.mockUserRepo.On("FindAllPaginated",
				ctx,
				mock.AnythingOfType("bson.M"),
				"created_at",
				-1,
				tc.page,
				tc.pageSize,
			).Return([]*entity.User{}, false, false, errors.New("test error"))

			err := s.service.Service(ctx, queryParam)

			s.Require().NoError(err)
			s.mockUserRepo.AssertExpectations(s.T())
		})
	}
}

func TestGetUsersTestSuite(t *testing.T) {
	suite.Run(t, new(getUsersTestSuite))
}
