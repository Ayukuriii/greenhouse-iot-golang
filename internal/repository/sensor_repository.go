package repository

import (
	"context"
	"greenhouse-iot-golang/internal/model"
)

type SensorRepository interface {
	Create(ctx context.Context, reading *model.SensorReading) error
}
