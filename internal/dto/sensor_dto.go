package dto

import "time"

type SensorDataRequest struct {
	DeviceID   string     `json:"device_id" validate:"required"`
	SensorType string     `json:"sensor_type" validate:"required,oneof=temperature humidity"`
	Value      *float64   `json:"value" validate:"required"`
	Unit       string     `json:"unit"`
	RecordedAt *time.Time `json:"recorded_at"`
}
