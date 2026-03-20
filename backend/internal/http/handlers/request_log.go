package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func requestID(c *fiber.Ctx) string {
	if value, ok := c.Locals("requestid").(string); ok {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "-"
}
