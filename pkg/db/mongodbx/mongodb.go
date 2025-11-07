package mongodbx

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Config struct {
	Username   string
	Password   string
	Host       string
	Port       int
	AuthSource *string
}

func NewMongoDB(ctx context.Context, config *Config) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/", config.Username, config.Password, config.Host, config.Port)
	if config.AuthSource != nil {
		uri += fmt.Sprintf("?authSource=%s", *config.AuthSource)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
