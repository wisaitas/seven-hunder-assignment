package initial

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/pkg/db/mongodbx"
	"github.com/7-solutions/backend-challenge/pkg/db/redisx"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type client struct {
	mongoDB *mongo.Client
	redis   redisx.Redis
}

func newClient() *client {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoDB, err := mongodbx.NewMongoDB(ctxTimeout, &mongodbx.Config{
		Username:   app.Config.MongoDB.Username,
		Password:   app.Config.MongoDB.Password,
		Host:       app.Config.MongoDB.Host,
		Port:       app.Config.MongoDB.Port,
		AuthSource: app.Config.MongoDB.AuthSource,
	})
	if err != nil {
		log.Fatalf("failed to connect to mongo db: %v", err)
	}

	redisClient, err := redisx.NewRedis(ctxTimeout, &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", app.Config.Redis.Host, app.Config.Redis.Port),
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	return &client{
		mongoDB: mongoDB,
		redis:   redisClient,
	}
}
