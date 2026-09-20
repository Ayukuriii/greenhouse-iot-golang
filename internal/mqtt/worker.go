package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"greenhouse-iot-golang/internal/config"
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/model"
	"greenhouse-iot-golang/internal/repository"
	"log/slog"
	"time"
)

type Worker struct {
	client     *Client
	deviceRepo repository.DeviceRepository
	commandCh  <-chan model.DeviceCommand
	cfg        *config.Config
}

func NewWorker(
	client *Client,
	deviceRepo repository.DeviceRepository,
	commandCh <-chan model.DeviceCommand,
	cfg *config.Config,
) *Worker {
	return &Worker{
		client:     client,
		deviceRepo: deviceRepo,
		commandCh:  commandCh,
		cfg:        cfg,
	}
}

// Start runs the worker loop in a dedicated background goroutine.
func (w *Worker) Start(ctx context.Context) {
	slog.Info("MQTT publisher worker goroutine started")

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("MQTT publisher worker received shutdown signal; stopping")
				return
			case cmd, ok := <-w.commandCh:
				if !ok {
					slog.Info("Command channel closed; worker exiting")
					return
				}
				w.processCommand(cmd)
			}
		}
	}()
}

func (w *Worker) processCommand(cmd model.DeviceCommand) {
	// Guard against unhandled panics within the worker to guarantee worker loop continuity
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic inside MQTT worker", "panic", r, "command_id", cmd.ID)
		}
	}()

	now := time.Now().UTC()
	topic := fmt.Sprintf("%s/%s", w.cfg.MQTTCommandTopicPrefix, cmd.DeviceID)

	payloadObj := dto.MQTTDevicePayload{
		DeviceID:  cmd.DeviceID,
		Command:   cmd.Command,
		Timestamp: now.Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		slog.Error("Failed to serialize MQTT payload", "error", err, "command_id", cmd.ID)
		_ = w.deviceRepo.UpdateCommandStatus(context.Background(), cmd.ID, model.CommandStatusFailed, nil)
		return
	}

	// Verify broker connection status
	if !w.client.IsConnected() {
		slog.Error("MQTT client disconnected; marking command as failed", "command_id", cmd.ID, "topic", topic)
		_ = w.deviceRepo.UpdateCommandStatus(context.Background(), cmd.ID, model.CommandStatusFailed, nil)
		return
	}

	// Publish with QoS 1
	token := w.client.Client.Publish(topic, 1, false, payloadBytes)
	if token.WaitTimeout(3*time.Second) && token.Error() != nil {
		slog.Error("Failed to publish command to MQTT topic", "error", token.Error(), "command_id", cmd.ID, "topic", topic)
		_ = w.deviceRepo.UpdateCommandStatus(context.Background(), cmd.ID, model.CommandStatusFailed, nil)
		return
	}

	slog.Info("Successfully published command to MQTT", "topic", topic, "command_id", cmd.ID)
	_ = w.deviceRepo.UpdateCommandStatus(context.Background(), cmd.ID, model.CommandStatusPublished, &now)
}
