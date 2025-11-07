package getuserbyid_test

import (
	"errors"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/getuserbyid"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type getUserByIDTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	service      getuserbyid.Service
	app          *fiber.App
}

func (s *getUserByIDTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.service = getuserbyid.NewService(s.mockUserRepo)
	s.app = fiber.New()
}

func (s *getUserByIDTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *getUserByIDTestSuite) TestGetUserByIDSuccess() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.MatchedBy(func(filter bson.M) bool {
			return filter["_id"] == testUserID && filter["deleted_at"] == nil
		}),
		mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = testUser
		})

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusOK, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *getUserByIDTestSuite) TestGetUserByIDNotFound() {
	testUserID := bson.NewObjectID()

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNotFound, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *getUserByIDTestSuite) TestGetUserByIDDatabaseError() {
	testUserID := bson.NewObjectID()

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(errors.New("database connection error"))

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *getUserByIDTestSuite) TestGetUserByIDDeletedUser() {
	testUserID := bson.NewObjectID()

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.MatchedBy(func(filter bson.M) bool {
			return filter["_id"] == testUserID && filter["deleted_at"] == nil
		}),
		mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNotFound, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func TestGetUserByIDTestSuite(t *testing.T) {
	suite.Run(t, new(getUserByIDTestSuite))
}
