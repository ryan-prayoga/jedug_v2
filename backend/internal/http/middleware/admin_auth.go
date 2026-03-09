package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

// AdminAuth returns a Fiber middleware that validates admin session tokens.
func AdminAuth(svc service.AdminService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return response.Error(c, fiber.StatusUnauthorized, "unauthorized")
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		session := svc.ValidateSession(token)
		if session == nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid or expired session")
		}

		c.Locals("admin_username", session.Username)
		return c.Next()
	}
}
