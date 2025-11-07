package initial

import appRepository "github.com/7-solutions/backend-challenge/internal/app/domain/repository"

type repository struct {
	userRepository appRepository.UserRepository
}

func newRepository(client *client) *repository {
	return &repository{
		userRepository: appRepository.NewUserRepository(client.mongoDB),
	}
}
