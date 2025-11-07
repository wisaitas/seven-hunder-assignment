package getusers

import (
	"github.com/7-solutions/backend-challenge/pkg/httpx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type QueryParam struct {
	httpx.PaginationQuery
}

type Response struct {
	ID    bson.ObjectID `json:"id"`
	Name  string        `json:"name"`
	Email string        `json:"email"`
}
