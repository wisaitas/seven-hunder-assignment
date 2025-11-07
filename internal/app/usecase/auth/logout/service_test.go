package logout_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/logout"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type logoutTestSuite struct {
	suite.Suite
	mockRedis *redisx.MockRedis
	service   logout.Service
	app       *fiber.App
}

func (s *logoutTestSuite) SetupTest() {
	s.mockRedis = redisx.NewMockRedis(s.T())
	s.service = logout.NewService(s.mockRedis)
	s.app = fiber.New()
}

func (s *logoutTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *logoutTestSuite) TestLogoutSuccess() {
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

	userContext := entity.UserContext{
		User: testUser,
	}

	accessTokenKey := fmt.Sprintf("access_token:%s", testUserID.Hex())
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", testUserID.Hex())

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == accessTokenKey
	})).Return(nil)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == refreshTokenKey
	})).Return(nil)

	err := s.service.Service(ctx, userContext)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *logoutTestSuite) TestLogoutAccessTokenDelError() {
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

	userContext := entity.UserContext{
		User: testUser,
	}

	accessTokenKey := fmt.Sprintf("access_token:%s", testUserID.Hex())

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == accessTokenKey
	})).Return(errors.New("redis connection error"))

	err := s.service.Service(ctx, userContext)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *logoutTestSuite) TestLogoutRefreshTokenDelError() {
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

	userContext := entity.UserContext{
		User: testUser,
	}

	accessTokenKey := fmt.Sprintf("access_token:%s", testUserID.Hex())
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", testUserID.Hex())

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == accessTokenKey
	})).Return(nil)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == refreshTokenKey
	})).Return(errors.New("redis connection error"))

	err := s.service.Service(ctx, userContext)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *logoutTestSuite) TestLogoutWithDifferentUserIDs() {
	testCases := []struct {
		name   string
		userID bson.ObjectID
	}{
		{
			name:   "User 1",
			userID: bson.NewObjectID(),
		},
		{
			name:   "User 2",
			userID: bson.NewObjectID(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			testUser := entity.User{
				ID:        tc.userID,
				Name:      tc.name,
				Email:     fmt.Sprintf("%s@example.com", tc.name),
				Password:  "hashedpassword",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			}

			userContext := entity.UserContext{
				User: testUser,
			}

			accessTokenKey := fmt.Sprintf("access_token:%s", tc.userID.Hex())
			refreshTokenKey := fmt.Sprintf("refresh_token:%s", tc.userID.Hex())

			ctx := s.createTestContext()
			defer s.app.ReleaseCtx(ctx)

			s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
				return len(keys) == 1 && keys[0] == accessTokenKey
			})).Return(nil)

			s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
				return len(keys) == 1 && keys[0] == refreshTokenKey
			})).Return(nil)

			err := s.service.Service(ctx, userContext)

			s.Require().NoError(err)
			s.Require().Equal(fiber.StatusNoContent, ctx.Response().StatusCode())
			s.mockRedis.AssertExpectations(s.T())
		})
	}
}

func TestLogoutTestSuite(t *testing.T) {
	suite.Run(t, new(logoutTestSuite))
}
