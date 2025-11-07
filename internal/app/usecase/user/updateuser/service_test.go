package updateuser_test

import (
	"errors"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/updateuser"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type updateUserTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	service      updateuser.Service
	app          *fiber.App
}

func (s *updateUserTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.service = updateuser.NewService(s.mockUserRepo)
	s.app = fiber.New()
}

func (s *updateUserTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *updateUserTestSuite) TestUpdateUserSuccessWithName() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Old Name",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	newName := "New Name"
	request := &updateuser.Request{
		Name: &newName,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = testUser
		})

	s.mockUserRepo.On("UpdateUser", ctx, testUserID, testUser.Version,
		mock.MatchedBy(func(updates bson.M) bool {
			return updates["name"] == newName && updates["updated_at"] != nil
		})).
		Return(nil)

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *updateUserTestSuite) TestUpdateUserSuccessWithEmail() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "old@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	newEmail := "new@example.com"
	request := &updateuser.Request{
		Email: &newEmail,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = testUser
		})

	s.mockUserRepo.On("FindByEmail", ctx, newEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			user.ID = testUserID
		})

	s.mockUserRepo.On("UpdateUser", ctx, testUserID, testUser.Version,
		mock.MatchedBy(func(updates bson.M) bool {
			return updates["email"] == newEmail
		})).
		Return(nil)

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *updateUserTestSuite) TestUpdateUserNotFound() {
	testUserID := bson.NewObjectID()
	newName := "New Name"
	request := &updateuser.Request{
		Name: &newName,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNotFound, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *updateUserTestSuite) TestUpdateUserEmailAlreadyExists() {
	testUserID := bson.NewObjectID()
	otherUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	newEmail := "existing@example.com"
	request := &updateuser.Request{
		Email: &newEmail,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = testUser
		})

	s.mockUserRepo.On("FindByEmail", ctx, newEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			user.ID = otherUserID
		})

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusConflict, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *updateUserTestSuite) TestUpdateUserNoFieldsToUpdate() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &updateuser.Request{}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = testUser
		})

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusBadRequest, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *updateUserTestSuite) TestUpdateUserDatabaseError() {
	testUserID := bson.NewObjectID()

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByID", ctx,
		mock.AnythingOfType("bson.M"),
		mock.AnythingOfType("*entity.User")).
		Return(errors.New("database connection error"))

	newName := "New Name"
	request := &updateuser.Request{
		Name: &newName,
	}

	err := s.service.Service(ctx, testUserID, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func TestUpdateUserTestSuite(t *testing.T) {
	suite.Run(t, new(updateUserTestSuite))
}
