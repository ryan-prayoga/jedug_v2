package middleware

import "github.com/gofiber/fiber/v2"

// NoStore disables intermediary/browser caching for sensitive endpoints.
func NoStore() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "no-store, no-cache, must-revalidate")
		c.Set(fiber.HeaderPragma, "no-cache")
		c.Set(fiber.HeaderExpires, "0")
		return c.Next()
	}
}
