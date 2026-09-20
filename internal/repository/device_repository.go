package repository

import (
	"context"
	"greenhouse-iot-golang/internal/model"
	"time"
)

type DeviceRepository interface {
	CreateCommand(ctx context.Context, cmd *model.DeviceCommand) error
	UpdateCommandStatus(ctx context.Context, id int64, status string, publishedAt *time.Time) error
}
