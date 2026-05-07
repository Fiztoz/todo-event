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
)

func registerRoutes(app *fiber.App) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	healthRepo := healthadapter.NewMongoRepository(clientIO)
	defer func() { healthRepo.Disconnect(context.Background()) }()

	healthService := application.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	app.Get("/health", healthHandler.CheckHealth)
}

func main() {
	app := fiber.New()
	registerRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
