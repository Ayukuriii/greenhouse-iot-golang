package model

import (
	"time"

	"github.com/google/uuid"
)

type SensorReading struct {
	ID         int64     `json:"-" db:"id"`
	PublicID   uuid.UUID `json:"public_id" db:"public_id"`
	DeviceID   string    `json:"device_id" db:"device_id"`
	SensorType string    `json:"sensor_type" db:"sensor_type"`
	Value      float64   `json:"value" db:"value"`
	Unit       string    `json:"unit" db:"unit"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
