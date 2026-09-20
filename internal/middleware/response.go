package middleware

import "github.com/gofiber/fiber/v3"

type SuccessEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Success bool         `json:"success"`
	Error   ErrorMessage `json:"error"`
}

type ErrorMessage struct {
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

func SendSuccess(c fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(SuccessEnvelope{
		Success: true,
		Data:    data,
	})
}

func SendError(c fiber.Ctx, statusCode int, message string, details []ErrorDetail) error {
	return c.Status(statusCode).JSON(ErrorEnvelope{
		Success: false,
		Error: ErrorMessage{
			Message: message,
			Details: details,
		},
	})
}
