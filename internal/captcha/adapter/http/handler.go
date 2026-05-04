package httpadapter

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/captcha/application"
	"todoe/internal/captcha/port"
)

type Handler struct {
	useCase port.UseCase
}

func NewHandler(useCase port.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Issue(c *fiber.Ctx) error {
	result := h.useCase.Issue(c.Context())
	if result.IsError() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	ch := result.MustGet()
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         ch.ID,
		"question":   ch.Question,
		"expires_at": ch.ExpiresAt,
	})
}

func (h *Handler) Verify(c *fiber.Ctx) error {
	id, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var body struct {
		Answer int `json:"answer"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := h.useCase.Verify(c.Context(), id, body.Answer)
	if result.IsError() {
		switch {
		case errors.Is(result.Error(), application.ErrChallengeNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": result.Error().Error()})
		case errors.Is(result.Error(), application.ErrChallengeExpired),
			errors.Is(result.Error(), application.ErrChallengeAlreadyUsed):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": result.Error().Error()})
		case errors.Is(result.Error(), application.ErrIncorrectAnswer):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": result.Error().Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"verified": true})
}
