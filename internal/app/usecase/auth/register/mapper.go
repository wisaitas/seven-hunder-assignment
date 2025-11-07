package register

import (
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *service) mapRequestToEntity(request *Request) *entity.User {
	return &entity.User{
		BaseEntity: entity.BaseEntity{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	}
}
