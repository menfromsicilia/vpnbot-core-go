package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// APIKeyMiddleware checks for valid API key in X-Api-Key header
func APIKeyMiddleware(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-Api-Key")
		if key == "" || key != apiKey {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden",
			})
		}
		return c.Next()
	}
}

