package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	CommandStatusPending   = "pending"
	CommandStatusPublished = "published"
	CommandStatusFailed    = "failed"
)

type DeviceCommand struct {
	ID          int64      `json:"-" db:"id"`
	PublicID    uuid.UUID  `json:"public_id" db:"public_id"`
	DeviceID    string     `json:"device_id" db:"device_id"`
	Command     string     `json:"command" db:"command"`
	Status      string     `json:"status" db:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
