package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit returns a Fiber middleware that limits requests per IP address.
// When the limit is exceeded, it logs the event and returns 429 with retry_after.
func RateLimit(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Printf("[ANTISPAM] rate_limit ip=%s path=%s method=%s", c.IP(), c.Path(), c.Method())
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"message":     "Terlalu banyak permintaan. Coba lagi nanti.",
				"retry_after": int(window.Seconds()),
			})
		},
	})
}
