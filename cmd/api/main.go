package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"greenhouse-iot-golang/internal/config"
	"greenhouse-iot-golang/internal/database"
	"greenhouse-iot-golang/internal/handler"
	"greenhouse-iot-golang/internal/middleware"
	"greenhouse-iot-golang/internal/model"
	"greenhouse-iot-golang/internal/mqtt"
	"greenhouse-iot-golang/internal/repository"
	"greenhouse-iot-golang/internal/service"
)

func main() {
	// 1. Initialize Structured Logger
	loggerHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(loggerHandler))

	slog.Info("Starting Greenhouse IoT Backend Service")

	// 2. Load Configuration
	cfg := config.LoadConfig()

	// 3. Initialize Database Connection
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Fatal: Unable to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Initialize MQTT Connection
	mqttClient, err := mqtt.NewMQTTClient(cfg)
	if err != nil {
		slog.Warn("MQTT broker not available on startup (will attempt auto-reconnect)", "error", err)
	}

	// 5. Initialize Validation
	validate := validator.New()

	// 6. Setup Channels & Concurrency Worker
	commandCh := make(chan model.DeviceCommand, cfg.CommandChannelBufferSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 7. Layer Dependency Injection
	sensorRepo := repository.NewSensorRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	sensorService := service.NewSensorService(sensorRepo)
	deviceService := service.NewDeviceService(deviceRepo, commandCh)

	sensorHandler := handler.NewSensorHandler(sensorService, validate)
	deviceHandler := handler.NewDeviceHandler(deviceService, validate)
	statusHandler := handler.NewStatusHandler(db, mqttClient)

	// 8. Start Background MQTT Worker
	mqttWorker := mqtt.NewWorker(mqttClient, deviceRepo, commandCh, cfg)
	mqttWorker.Start(ctx)

	// 9. Initialize Fiber v3 Application
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler(),
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New())

	// Routes
	app.Post("/sensor-data", sensorHandler.CreateSensorData)
	app.Post("/device-control", deviceHandler.ControlDevice)
	app.Get("/status", statusHandler.GetStatus)

	// 10. Graceful Shutdown Routine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		slog.Info("HTTP Server listening", "port", cfg.Port)
		if err := app.Listen(addr); err != nil {
			slog.Info("HTTP server stopped listening", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Initiating graceful shutdown...")
	cancel() // Stop background workers

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		slog.Error("Error during HTTP server shutdown", "error", err)
	}

	slog.Info("Server shutdown completed cleanly")
}
