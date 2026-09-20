package service

import (
	"context"
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/model"
	"greenhouse-iot-golang/internal/repository"
	"time"
)

type sensorServiceImpl struct {
	repo repository.SensorRepository
}

func NewSensorService(repo repository.SensorRepository) SensorService {
	return &sensorServiceImpl{repo: repo}
}

func (s *sensorServiceImpl) RecordSensorData(ctx context.Context, req *dto.SensorDataRequest) (*model.SensorReading, error) {
	recordedAt := time.Now().UTC()
	if req.RecordedAt != nil && !req.RecordedAt.IsZero() {
		recordedAt = req.RecordedAt.UTC()
	}

	reading := &model.SensorReading{
		DeviceID:   req.DeviceID,
		SensorType: req.SensorType,
		Value:      *req.Value,
		Unit:       req.Unit,
		RecordedAt: recordedAt,
	}

	if err := s.repo.Create(ctx, reading); err != nil {
		return nil, err
	}

	return reading, nil
}
