package controller

import "github.com/gofiber/fiber/v2"

func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok", "service": "retailpulse-api"})
	})
}
