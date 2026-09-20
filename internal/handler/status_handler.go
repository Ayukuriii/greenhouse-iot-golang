package handler

import (
	"context"
	"greenhouse-iot-golang/internal/middleware"
	"greenhouse-iot-golang/internal/mqtt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
)

type StatusHandler struct {
	db         *sqlx.DB
	mqttClient *mqtt.Client
}

func NewStatusHandler(db *sqlx.DB, mqttClient *mqtt.Client) *StatusHandler {
	return &StatusHandler{
		db:         db,
		mqttClient: mqttClient,
	}
}

func (h *StatusHandler) GetStatus(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "disconnected"
	}

	mqttStatus := "connected"
	if !h.mqttClient.IsConnected() {
		mqttStatus = "disconnected"
	}

	overallStatus := "ok"
	if dbStatus != "connected" || mqttStatus != "connected" {
		overallStatus = "degraded"
	}

	statusData := fiber.Map{
		"status":    overallStatus,
		"service":   "up",
		"database":  dbStatus,
		"mqtt":      mqttStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return middleware.SendSuccess(c, fiber.StatusOK, statusData)
}
