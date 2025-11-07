package createuser

import (
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *service) mapRequestToEntity(request *Request) *entity.User {
	return &entity.User{
		ID:        bson.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      request.Name,
		Email:     request.Email,
		Password:  request.Password,
		Version:   1,
	}
}
