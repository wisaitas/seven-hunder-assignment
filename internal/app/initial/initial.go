package initial

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/7-solutions/backend-challenge/internal/app"
	"github.com/7-solutions/backend-challenge/internal/app/domain/entity"
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func init() {
	if err := env.Parse(&app.Config); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}
}

type App struct {
	FiberApp *fiber.App
	Client   *client
}

func New() *App {
	client := newClient()
	sdk := newSdk()
	app := fiber.New(
		fiber.Config{
			BodyLimit:               app.Config.Service.BodyLimit * 1024 * 1024,
			EnableTrustedProxyCheck: true,
			TrustedProxies: []string{
				"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
			},
			ProxyHeader:  fiber.HeaderXForwardedFor,
			ReadTimeout:  app.Config.Service.ReadTimeout,
			WriteTimeout: app.Config.Service.WriteTimeout,
		},
	)

	newMiddleware(app, client)
	repository := newRepository(client)
	useCase := newUseCase(client, repository, sdk)
	newRouter(app, useCase)

	return &App{
		FiberApp: app,
		Client:   client,
	}
}

func (i *App) Run() {
	go func() {
		i.backgroundProcess()
	}()

	go func() {
		if err := i.FiberApp.Listen(fmt.Sprintf(":%d", app.Config.Service.Port)); err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func (i *App) Close() {
	if err := i.Client.mongoDB.Disconnect(context.Background()); err != nil {
		log.Fatalf("failed to disconnect mongo db: %v", err)
	}

	if err := i.Client.redis.Client().Close(); err != nil {
		log.Fatalf("failed to close redis: %v", err)
	}

	if err := i.FiberApp.Shutdown(); err != nil {
		log.Fatalf("failed to shutdown fiber app: %v", err)
	}

	log.Println("gracefully shutdown app")
}

func (i *App) backgroundProcess() {
	for {
		time.Sleep(10 * time.Second)

		count, err := i.Client.mongoDB.Database(app.Config.MongoDB.Database).Collection(entity.User{}.CollectionName()).
			CountDocuments(context.Background(), bson.M{"deleted_at": nil})
		if err != nil {
			log.Fatalf("failed to count users: %v", err)
		}
		log.Printf("number of users: %d", count)
	}
}
