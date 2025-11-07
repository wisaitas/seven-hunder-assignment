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

	// Set up default config for testing
	app.Config.JWT.Secret = "test-secret-key-32-bytes-long!" // 32 bytes for AES-256
	app.Config.JWT.AccessTTL = 15                            // 15 minutes
	app.Config.JWT.RefreshTTL = 24                           // 24 hours
}

// Helper function to create fiber context for testing
func (s *loginTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *loginTestSuite) TestLoginSuccess() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Del for attempts key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	// Mock Redis Del for block count key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	// Mock JWT Generate for access token
	s.mockJwt.On("Generate",
		mock.MatchedBy(func(claims interface{}) bool {
			ec, ok := claims.(entity.ExternalContext)
			return ok && ec.Subject == testUser.ID.Hex()
		}),
		app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	// Mock JWT Generate for refresh token
	s.mockJwt.On("Generate",
		mock.MatchedBy(func(claims interface{}) bool {
			ec, ok := claims.(entity.ExternalContext)
			return ok && ec.Subject == testUser.ID.Hex()
		}),
		app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	// Mock Redis Set for access token
	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(nil)

	// Mock Redis Set for refresh token
	s.mockRedis.On("Set", mock.Anything,
		refreshTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
	s.mockJwt.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginUserNotFound() {
	// Arrange
	testEmail := "notfound@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return not found error
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	// Mock Redis Get for attempts
	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	// Mock Redis Set for incrementing attempts
	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusNotFound, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginInvalidPassword() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Get for attempts (first attempt)
	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	// Mock Redis Set for incrementing attempts
	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusUnauthorized, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginMultipleFailedAttempts() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Get for attempts (already 2 attempts)
	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("2", nil)

	// Mock Redis Set for incrementing attempts to 3
	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"3",
		15*time.Minute).Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusUnauthorized, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRepositoryError() {
	// Arrange
	testEmail := "test@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return database error
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(errors.New("database connection error"))

	// Mock Redis Get for attempts
	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	// Mock Redis Set for incrementing attempts
	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRedisSetError() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Del for attempts key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	// Mock Redis Del for block count key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	// Mock JWT Generate for access token
	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	// Mock JWT Generate for refresh token (generated before Redis Set is called)
	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	// Mock Redis Set for access token - return error
	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).Return(errors.New("redis connection error"))

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
	s.mockJwt.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginRedisDelError() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Del for attempts key - return error
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(errors.New("redis del error"))

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginIncrementFailedAttemptsRedisError() {
	// Arrange
	testEmail := "test@example.com"
	testPassword := "password123"

	request := &login.Request{
		Email:    testEmail,
		Password: testPassword,
	}

	attemptsKey := fmt.Sprintf("login:attempts:%s", testEmail)

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return not found error
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	// Mock Redis Get for attempts
	s.mockRedis.On("Get", mock.Anything, attemptsKey).
		Return("", errors.New("redis: nil"))

	// Mock Redis Set for incrementing attempts - return error
	s.mockRedis.On("Set", mock.Anything,
		attemptsKey,
		"1",
		15*time.Minute).Return(errors.New("redis set error"))

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRedis.AssertExpectations(s.T())
}

func (s *loginTestSuite) TestLoginValidateTokenContent() {
	// Arrange
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

	// Create test context
	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	// Mock FindByEmail - return user
	s.mockUserRepo.On("FindByEmail", ctx, testEmail, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *testUser
		})

	// Mock Redis Del for attempts key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == attemptsKey
	})).Return(nil)

	// Mock Redis Del for block count key
	s.mockRedis.On("Del", mock.Anything, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == blockCountKey
	})).Return(nil)

	// Mock JWT Generate - return tokens
	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-access-token", nil).Once()

	s.mockJwt.On("Generate", mock.Anything, app.Config.JWT.Secret).
		Return("encrypted-refresh-token", nil).Once()

	// Mock Redis Set for access token - capture value
	s.mockRedis.On("Set", mock.Anything,
		accessTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).
		Run(func(args mock.Arguments) {
			capturedAccessTokenValue = args.Get(2).(string)
		}).
		Return(nil)

	// Mock Redis Set for refresh token - capture value
	s.mockRedis.On("Set", mock.Anything,
		refreshTokenKey,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration")).
		Run(func(args mock.Arguments) {
			capturedRefreshTokenValue = args.Get(2).(string)
		}).
		Return(nil)

	// Act
	err := s.service.Service(ctx, request)

	// Assert
	s.Require().NoError(err)
	s.Require().NotEmpty(capturedAccessTokenValue)
	s.Require().NotEmpty(capturedRefreshTokenValue)

	// Verify the stored user context can be unmarshaled
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
