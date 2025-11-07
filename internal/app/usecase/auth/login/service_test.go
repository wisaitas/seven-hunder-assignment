package login_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/login"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type loginTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	mockRedis    *redisx.MockRedis
	mockJwt      *jwtx.MockJwt
	service      login.Service
	app          *fiber.App
}

func (s *loginTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.mockRedis = redisx.NewMockRedis(s.T())
	s.mockJwt = jwtx.NewMockJwt(s.T())
	s.service = login.NewService(s.mockUserRepo, s.mockRedis, s.mockJwt)
	s.app = fiber.New()

	app.Config.JWT.Secret = "test-secret-key-32-bytes-long!"
	app.Config.JWT.AccessTTL = 15
	app.Config.JWT.RefreshTTL = 24
}

func (s *loginTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *loginTestSuite) TestLoginSuccess() {
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)
	blockCountKey := fmt.Sprintf("login:block_count:%s", testEmail)
	accessTokenKey := fmt.Sprintf("access_token:%s", testUser.ID.Hex())
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", testUser.ID.Hex())

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	s.mockJwt.On("Generate",
		mock.MatchedBy(func(claims interface{}) bool {
			ec, ok := claims.(entity.ExternalContext)
			return ok && ec.Subject == testUser.ID.Hex()
		}),
		app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	s.mockJwt.On("Generate",
		mock.MatchedBy(func(claims interface{}) bool {
			ec, ok := claims.(entity.ExternalContext)
			return ok && ec.Subject == testUser.ID.Hex()
		}),
		app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(nil)

	s.mockRedis.On("Set", mock.Anything,
		refreshTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
	s.mockJwt.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginUserNotFound() {
	testEmail := "notfound@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNotFound, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginInvalidPassword() {
	testEmail := "test@example.com"
	testPassword := "password123"
	wrongPassword := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: wrongPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusUnauthorized, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginMultipleFailedAttempts() {
	testEmail := "test@example.com"
	testPassword := "password123"
	wrongPassword := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: wrongPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("2", nil)

	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"3",
		15*time.Minute).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusUnauthorized, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRepositoryError() {
	testEmail := "test@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(errors.New("database connection error"))

	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRedisSetError() {
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)
	blockCountKey := fmt.Sprintf("login:block_count:%s", testEmail)
	accessTokenKey := fmt.Sprintf("access_token:%s", testUser.ID.Hex())

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(errors.New("redis connection error"))

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
	s.mockJwt.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRedisDelError() {
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(errors.New("redis del error"))

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginIncrementFailedAttemptsRedisError() {
	testEmail := "test@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(errors.New("redis set error"))

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginValidateTokenContent() {
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	testUser := &entity.User{
		ID:        bson.NewObjectID(),
		Name:      "Test User",
		Email:     testEmail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)
	blockCountKey := fmt.Sprintf("login:block_count:%s", testEmail)
	accessTokenKey := fmt.Sprintf("access_token:%s", testUser.ID.Hex())
	refreshTokenKey := fmt.Sprintf("refresh_token:%s", testUser.ID.Hex())

	var capturedAccessTokenValue string
	var capturedRefreshTokenValue string

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).
		Run(func(args mock.Arguments) {
			capturedAccessTokenValue = args.Get(2).(string)
		}).
		Return(nil)

	s.mockRedis.On("Set", mock.Anything,
		refreshTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).
		Run(func(args mock.Arguments) {
			capturedRefreshTokenValue = args.Get(2).(string)
		}).
		Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().NotEmpty(capturedAccessTokenValue)
	s.Require().NotEmpty(capturedRefreshTokenValue)

	var userContext entity.UserContext
	err = json.Unmarshal([]byte(capturedAccessTokenValue), &userContext)
	s.Require().NoError(err)
	s.Require().Equal(testUser.ID, userContext.User.ID)
	s.Require().Equal(testUser.Email, userContext.User.Email)
	s.Require().Equal(testUser.Name, userContext.User.Name)

	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
	s.mockJwt.AssertExpectations(s.T())
}

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(loginTestSuite))
}
