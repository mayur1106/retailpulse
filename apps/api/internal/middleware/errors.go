package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"retailpulse/apps/api/internal/domain"
)

func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "internal server error"

		var fiberErr *fiber.Error
		switch {
		case errors.As(err, &fiberErr):
			status = fiberErr.Code
			message = fiberErr.Message
		case errors.Is(err, domain.ErrValidation):
			status = fiber.StatusBadRequest
			message = err.Error()
		case errors.Is(err, domain.ErrConfiguration):
			status = fiber.StatusServiceUnavailable
			message = err.Error()
		case errors.Is(err, domain.ErrConflict):
			status = fiber.StatusConflict
			message = "resource already exists"
		case errors.Is(err, domain.ErrInvalidCredential), errors.Is(err, domain.ErrInvalidToken):
			status = fiber.StatusUnauthorized
			message = err.Error()
		case errors.Is(err, domain.ErrForbidden):
			status = fiber.StatusForbidden
			message = "forbidden"
		case errors.Is(err, domain.ErrNotFound):
			status = fiber.StatusNotFound
			message = "resource not found"
		default:
			logger.Error("request failed", "error", err)
		}

		return ctx.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"message": message,
				"status":  status,
			},
		})
	}
}
