package handler

import (
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/middleware"
	"greenhouse-iot-golang/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type SensorHandler struct {
	service  service.SensorService
	validate *validator.Validate
}

func NewSensorHandler(service service.SensorService, validate *validator.Validate) *SensorHandler {
	return &SensorHandler{
		service:  service,
		validate: validate,
	}
}

func (h *SensorHandler) CreateSensorData(c fiber.Ctx) error {
	var req dto.SensorDataRequest

	if err := c.Bind().Body(&req); err != nil {
		return middleware.SendError(c, fiber.StatusBadRequest, "Invalid JSON payload format", nil)
	}

	if err := h.validate.Struct(&req); err != nil {
		var details []middleware.ErrorDetail
		if valErrs, ok := err.(validator.ValidationErrors); ok {
			for _, v := range valErrs {
				details = append(details, middleware.ErrorDetail{
					Field:   v.Field(),
					Message: "Validation failed on tag: " + v.Tag(),
				})
			}
		}
		return middleware.SendError(c, fiber.StatusBadRequest, "Validation error", details)
	}

	reading, err := h.service.RecordSensorData(c.Context(), &req)
	if err != nil {
		return middleware.SendError(c, fiber.StatusInternalServerError, "Failed to record sensor reading", nil)
	}

	return middleware.SendSuccess(c, fiber.StatusCreated, reading)
}
