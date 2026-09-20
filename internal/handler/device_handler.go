package handler

import (
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/middleware"
	"greenhouse-iot-golang/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type DeviceHandler struct {
	service  service.DeviceService
	validate *validator.Validate
}

func NewDeviceHandler(service service.DeviceService, validate *validator.Validate) *DeviceHandler {
	return &DeviceHandler{
		service:  service,
		validate: validate,
	}
}

func (h *DeviceHandler) ControlDevice(c fiber.Ctx) error {
	var req dto.DeviceControlRequest

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

	cmd, err := h.service.SendDeviceCommand(c.Context(), &req)
	if err != nil {
		return middleware.SendError(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return middleware.SendSuccess(c, fiber.StatusOK, cmd)
}
