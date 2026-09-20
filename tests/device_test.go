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

type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) CreateCommand(ctx context.Context, cmd *model.DeviceCommand) error {
	args := m.Called(ctx, cmd)
	cmd.ID = 1
	cmd.PublicID = uuid.New()
	cmd.CreatedAt = time.Now().UTC()
	return args.Error(0)
}

func (m *MockDeviceRepository) UpdateCommandStatus(ctx context.Context, id int64, status string, publishedAt *time.Time) error {
	args := m.Called(ctx, id, status, publishedAt)
	return args.Error(0)
}

func setupDeviceApp(mockRepo *MockDeviceRepository) (*fiber.App, chan model.DeviceCommand) {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler(),
	})
	validate := validator.New()
	commandCh := make(chan model.DeviceCommand, 10)
	deviceService := service.NewDeviceService(mockRepo, commandCh)
	deviceHandler := handler.NewDeviceHandler(deviceService, validate)

	app.Post("/device-control", deviceHandler.ControlDevice)
	return app, commandCh
}

func TestControlDevice_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	mockRepo.On("CreateCommand", mock.Anything, mock.AnythingOfType("*model.DeviceCommand")).Return(nil)

	app, commandCh := setupDeviceApp(mockRepo)

	payload := map[string]interface{}{
		"device_id": "fan-001",
		"command":   "ON",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/device-control", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify command was pushed onto channel
	select {
	case cmd := <-commandCh:
		assert.Equal(t, "fan-001", cmd.DeviceID)
		assert.Equal(t, "ON", cmd.Command)
	default:
		t.Fatal("Expected command to be in channel, but channel was empty")
	}
}

func TestControlDevice_ValidationFailure(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	app, _ := setupDeviceApp(mockRepo)

	// Invalid command value (must be ON or OFF)
	payload := map[string]interface{}{
		"device_id": "fan-001",
		"command":   "PAUSE",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/device-control", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
