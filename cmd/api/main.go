package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	"todoe/internal/health/application"
	pkgio "todoe/pkg/io"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}

	clientIO := pkgio.Lazy(func() mo.Result[*mongo.Client] {
		client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
		if err != nil {
			return mo.Err[*mongo.Client](err)
		}
		return mo.Ok(client)
	})

	defer func() {
		if clientIO.IsInitialized() {
			if r := clientIO.Run(); r.IsOk() {
				_ = r.MustGet().Disconnect(context.Background())
			}
		}
	}()

	healthRepo := healthadapter.NewMongoRepository(&clientIO)
	healthService := application.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)

	log.Fatal(app.Listen(":3000"))
}
