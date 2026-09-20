package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func CustomErrorHandler() fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		}

		slog.Error("HTTP request failed",
			"status", code,
			"method", c.Method(),
			"path", c.Path(),
			"error", err.Error(),
		)

		return SendError(c, code, err.Error(), nil)
	}
}
