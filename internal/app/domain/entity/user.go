package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `bson:"_id"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
	DeletedAt *time.Time    `bson:"deleted_at"`
	Name      string        `bson:"name"`
	Email     string        `bson:"email"`
	Password  string        `bson:"password"`
	Version   int           `bson:"version"`
}

func (User) CollectionName() string {
	return "users"
}
