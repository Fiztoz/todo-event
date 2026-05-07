package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	captchaadapter "todoe/internal/captcha/adapter"
	captchahttp "todoe/internal/captcha/adapter/http"
	captchaapp "todoe/internal/captcha/application"
	captchadomain "todoe/internal/captcha/domain"
	"todoe/internal/event"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}
	port := os.Getenv("CAPTCHA_PORT")
	if port == "" {
		port = "3010"
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	bus := event.NewEventBus()
	repo := captchaadapter.NewMongoRepository(clientIO)
	projection := captchaadapter.NewProjectionHandler(repo)
	bus.Subscribe(captchadomain.EventIssued, projection)
	bus.Subscribe(captchadomain.EventVerified, projection)

	service := captchaapp.NewService(repo, bus)
	handler := captchahttp.NewHandler(service)

	app := fiber.New()
	app.Post("/captcha", handler.Issue)
	app.Post("/captcha/:id/verify", handler.Verify)

	slog.Info("captcha service listening", "port", port)
	log.Fatal(app.Listen(":" + port))
}
