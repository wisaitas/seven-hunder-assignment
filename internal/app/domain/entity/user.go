package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time    `bson:"deleted_at" json:"deleted_at"`
	Name      string        `bson:"name" json:"name"`
	Email     string        `bson:"email" json:"email"`
	Password  string        `bson:"password" json:"password"`
	Version   int           `bson:"version" json:"version"`
}

func (User) CollectionName() string {
	return "users"
}
