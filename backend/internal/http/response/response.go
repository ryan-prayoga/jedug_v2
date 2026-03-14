package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success   bool        `json:"success"`
	ErrorCode string      `json:"error_code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func OKMessage(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
	})
}

func Error(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Message: message,
	})
}

// ErrorWithCode returns a structured error response with a stable error_code that
// the frontend can map to a user-friendly message.
func ErrorWithCode(c *fiber.Ctx, statusCode int, errorCode, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success:   false,
		ErrorCode: errorCode,
		Message:   message,
	})
}
