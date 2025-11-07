package register_test

import (
	"errors"
	"testing"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/7-solutions/backend-challenge/internal/app/domain/repository"
	"github.com/7-solutions/backend-challenge/internal/app/usecase/auth/register"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type registerTestSuite struct {
	suite.Suite
	mockUserRepo *repository.MockUserRepository
	service      register.Service
	app          *fiber.App
}

func (s *registerTestSuite) SetupTest() {
	s.mockUserRepo = repository.NewMockUserRepository(s.T())
	s.service = register.NewService(s.mockUserRepo)
	s.app = fiber.New()
}

func (s *registerTestSuite) createTestContext() *fiber.Ctx {
	ctx := s.app.AcquireCtx(&fasthttp.RequestCtx{})
	return ctx
}

func (s *registerTestSuite) TestRegisterSuccess() {
	request := &register.Request{
		Name:            "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockUserRepo.On("CreateUser", ctx, mock.MatchedBy(func(user *entity.User) bool {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		return user.Name == request.Name &&
			user.Email == request.Email &&
			err == nil &&
			user.Version == 1
	})).Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusCreated, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *registerTestSuite) TestRegisterEmailAlreadyExists() {
	request := &register.Request{
		Name:            "Test User",
		Email:           "existing@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	existingUser := &entity.User{
		Name:  "Existing User",
		Email: request.Email,
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(2).(*entity.User)
			*user = *existingUser
		})

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusConflict, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *registerTestSuite) TestRegisterFindByEmailDatabaseError() {
	request := &register.Request{
		Name:            "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(errors.New("database connection error"))

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *registerTestSuite) TestRegisterCreateUserError() {
	request := &register.Request{
		Name:            "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).
		Return(errors.New("failed to create user"))

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusInternalServerError, ctx.Response().StatusCode())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *registerTestSuite) TestRegisterWithDifferentEmails() {
	testCases := []struct {
		name  string
		email string
	}{
		{
			name:  "User with Gmail",
			email: "user1@gmail.com",
		},
		{
			name:  "User with Outlook",
			email: "user2@outlook.com",
		},
		{
			name:  "User with Corporate Email",
			email: "user3@company.com",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			request := &register.Request{
				Name:            tc.name,
				Email:           tc.email,
				Password:        "password123",
				ConfirmPassword: "password123",
			}

			ctx := s.createTestContext()
			defer s.app.ReleaseCtx(ctx)

			s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
				Return(mongo.ErrNoDocuments)

			s.mockUserRepo.On("CreateUser", ctx, mock.MatchedBy(func(user *entity.User) bool {
				return user.Email == tc.email
			})).Return(nil)

			err := s.service.Service(ctx, request)

			s.Require().NoError(err)
			s.Require().Equal(fiber.StatusCreated, ctx.Response().StatusCode())
			s.mockUserRepo.AssertExpectations(s.T())
		})
	}
}

func (s *registerTestSuite) TestRegisterPasswordIsHashed() {
	plainPassword := "mySecretPassword123"
	request := &register.Request{
		Name:            "Test User",
		Email:           "test@example.com",
		Password:        plainPassword,
		ConfirmPassword: plainPassword,
	}

	var capturedUser *entity.User

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).
		Run(func(args mock.Arguments) {
			capturedUser = args.Get(1).(*entity.User)
		}).
		Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusCreated, ctx.Response().StatusCode())
	s.Require().NotNil(capturedUser)

	s.Require().NotEqual(plainPassword, capturedUser.Password)

	err = bcrypt.CompareHashAndPassword([]byte(capturedUser.Password), []byte(plainPassword))
	s.Require().NoError(err, "Hashed password should match original password")

	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *registerTestSuite) TestRegisterUserEntityFieldsAreSet() {
	request := &register.Request{
		Name:            "John Doe",
		Email:           "john.doe@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	var capturedUser *entity.User

	ctx := s.createTestContext()
	defer s.app.ReleaseCtx(ctx)

	s.mockUserRepo.On("FindByEmail", ctx, request.Email, mock.AnythingOfType("*entity.User")).
		Return(mongo.ErrNoDocuments)

	s.mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).
		Run(func(args mock.Arguments) {
			capturedUser = args.Get(1).(*entity.User)
		}).
		Return(nil)

	err := s.service.Service(ctx, request)

	s.Require().NoError(err)
	s.Require().NotNil(capturedUser)

	s.Require().NotEmpty(capturedUser.ID, "ID should be set")
	s.Require().Equal(request.Name, capturedUser.Name)
	s.Require().Equal(request.Email, capturedUser.Email)
	s.Require().NotEmpty(capturedUser.Password, "Password should be hashed and set")
	s.Require().NotZero(capturedUser.CreatedAt, "CreatedAt should be set")
	s.Require().NotZero(capturedUser.UpdatedAt, "UpdatedAt should be set")
	s.Require().Equal(1, capturedUser.Version)

	s.mockUserRepo.AssertExpectations(s.T())
}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(registerTestSuite))
}
