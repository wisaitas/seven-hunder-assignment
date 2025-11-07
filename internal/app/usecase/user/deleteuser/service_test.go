package deleteuser_test

import (
	"errors"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/user/deleteuser"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type deleteUserTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	service      deleteuser.Service
	app          *fiber.App
}

func (s *deleteUserTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.service = deleteuser.NewService(s.mockUserRepo)
	s.app = fiber.New()
}

func (s *deleteUserTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *deleteUserTestSuite) TestDeleteUserSuccess() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
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

	s.mockUserRepo.On("UpdateUser", ctx, testUserID, testUser.Version,
		mock.MatchedBy(func(updates bson.M) bool {
			_, hasDeletedAt := updates["deleted_at"]
			return hasDeletedAt
		})).
		Return(nil)

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *deleteUserTestSuite) TestDeleteUserNotFound() {
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

func (s *deleteUserTestSuite) TestDeleteUserFindByIDError() {
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

func (s *deleteUserTestSuite) TestDeleteUserUpdateError() {
	testUserID := bson.NewObjectID()
	testUser := entity.User{
		ID:        testUserID,
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
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

	s.mockUserRepo.On("UpdateUser", ctx, testUserID, testUser.Version, mock.AnythingOfType("bson.M")).
		Return(errors.New("update failed"))

	err := s.service.Service(ctx, testUserID)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *deleteUserTestSuite) TestDeleteUserAlreadyDeleted() {
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

func (s *deleteUserTestSuite) TestDeleteMultipleUsers() {
	testCases := []struct {
		name   string
		userID bson.ObjectID
		email  string
	}{
		{
			name:   "Delete User 1",
			userID: bson.NewObjectID(),
			email:  "user1@example.com",
		},
		{
			name:   "Delete User 2",
			userID: bson.NewObjectID(),
			email:  "user2@example.com",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			testUser := entity.User{
				ID:        tc.userID,
				Name:      tc.name,
				Email:     tc.email,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
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

			s.mockUserRepo.On("UpdateUser", ctx, tc.userID, testUser.Version, mock.AnythingOfType("bson.M")).
				Return(nil)

			err := s.service.Service(ctx, tc.userID)

			s.Require().NoError(err)
			s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
			s.mockUserRepo.AssertExpectations(s.T())
		})
	}
}

func TestDeleteUserTestSuite(t *testing.T) {
	suite.Run(t, new(deleteUserTestSuite))
}
