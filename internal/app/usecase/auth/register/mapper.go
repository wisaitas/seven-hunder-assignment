package register

import (
	"time"

	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *service) mapRequestToEntity(request *Request) *entity.User {
	timeNow := time.Now()
	return &entity.User{
		ID:        bson.NewObjectID(),
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      request.Name,
		Email:     request.Email,
		Password:  request.Password,
		Version:   1,
	}
}
