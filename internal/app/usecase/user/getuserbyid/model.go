package getuserbyid

import "go.mongodb.org/mongo-driver/v2/bson"

type Response struct {
	ID    bson.ObjectID `json:"id"`
	Name  string        `json:"name"`
	Email string        `json:"email"`
}
