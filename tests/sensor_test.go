package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"greenhouse-iot-golang/internal/handler"
	"greenhouse-iot-golang/internal/middleware"
	"greenhouse-iot-golang/internal/model"
	"greenhouse-iot-golang/internal/service"
)

type MockSensorRepository struct {
	mock.Mock
}

func (m *MockSensorRepository) Create(ctx context.Context, reading *model.SensorReading) error {
	args := m.Called(ctx, reading)
	reading.ID = 1
	reading.PublicID = uuid.New()
	reading.CreatedAt = time.Now().UTC()
	return args.Error(0)
}

func setupSensorApp(mockRepo *MockSensorRepository) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler(),
	})
	validate := validator.New()
	sensorService := service.NewSensorService(mockRepo)
	sensorHandler := handler.NewSensorHandler(sensorService, validate)

	app.Post("/sensor-data", sensorHandler.CreateSensorData)
	return app
}

func TestCreateSensorData_Success(t *testing.T) {
	mockRepo := new(MockSensorRepository)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.SensorReading")).Return(nil)

	app := setupSensorApp(mockRepo)

	payload := map[string]interface{}{
		"device_id":   "sensor-001",
		"sensor_type": "temperature",
		"value":       27.5,
		"unit":        "celsius",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/sensor-data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateSensorData_ValidationFailure(t *testing.T) {
	mockRepo := new(MockSensorRepository)
	app := setupSensorApp(mockRepo)

	// Missing 'value' and invalid 'sensor_type'
	payload := map[string]interface{}{
		"device_id":   "sensor-001",
		"sensor_type": "invalid_type",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/sensor-data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
