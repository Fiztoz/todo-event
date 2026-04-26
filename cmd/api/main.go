package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	"todoe/internal/health/application"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	healthRepo := healthadapter.NewMongoRepository(mongoClient)
	healthService := application.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)

	log.Fatal(app.Listen(":3000"))
}
