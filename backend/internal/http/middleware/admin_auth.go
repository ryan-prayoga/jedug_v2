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
		token := strings.TrimSpace(c.Cookies(service.AdminSessionCookieName))
		if token == "" {
			auth := c.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}
		if token == "" {
			return response.Error(c, fiber.StatusUnauthorized, "unauthorized")
		}

		session := svc.ValidateSession(token)
		if session == nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid or expired session")
		}

		c.Locals("admin_username", session.Username)
		c.Locals("admin_session_token", token)
		return c.Next()
	}
}
