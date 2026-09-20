package service

import (
	"context"
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/model"
)

type DeviceService interface {
	SendDeviceCommand(ctx context.Context, req *dto.DeviceControlRequest) (*model.DeviceCommand, error)
}
