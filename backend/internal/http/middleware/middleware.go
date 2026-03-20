package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func Register(app *fiber.App, corsOrigins string) {
	allowCredentials := hasExplicitOrigins(corsOrigins)

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${method} ${path} ${status} ${latency} rid=${reqHeader:X-Request-Id}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, X-Device-Token, X-Upload-Token, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: allowCredentials,
	}))
}

func hasExplicitOrigins(corsOrigins string) bool {
	for _, origin := range strings.Split(corsOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			return false
		}
	}

	return strings.TrimSpace(corsOrigins) != ""
}
