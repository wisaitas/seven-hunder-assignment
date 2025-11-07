package getusers

import (
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QueryParam struct {
	httpx.PaginationQuery
}

type Response struct {
	ID    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Email string             `json:"email"`
}
