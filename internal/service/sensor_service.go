package service

import (
	"context"
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/model"
)

type SensorService interface {
	RecordSensorData(ctx context.Context, req *dto.SensorDataRequest) (*model.SensorReading, error)
}
