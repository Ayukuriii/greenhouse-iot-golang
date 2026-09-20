package service

import (
	"context"
	"fmt"
	"greenhouse-iot-golang/internal/dto"
	"greenhouse-iot-golang/internal/model"
	"greenhouse-iot-golang/internal/repository"
)

type deviceServiceImpl struct {
	repo      repository.DeviceRepository
	commandCh chan<- model.DeviceCommand
}

func NewDeviceService(repo repository.DeviceRepository, commandCh chan<- model.DeviceCommand) DeviceService {
	return &deviceServiceImpl{
		repo:      repo,
		commandCh: commandCh,
	}
}

func (s *deviceServiceImpl) SendDeviceCommand(ctx context.Context, req *dto.DeviceControlRequest) (*model.DeviceCommand, error) {
	cmd := &model.DeviceCommand{
		DeviceID: req.DeviceID,
		Command:  req.Command,
		Status:   model.CommandStatusPending,
	}

	// Persist initial command status in DB
	if err := s.repo.CreateCommand(ctx, cmd); err != nil {
		return nil, err
	}

	// Non-blocking enqueue to Go channel
	select {
	case s.commandCh <- *cmd:
		// Successfully queued
	default:
		return nil, fmt.Errorf("command channel queue is full; please retry later")
	}

	return cmd, nil
}
