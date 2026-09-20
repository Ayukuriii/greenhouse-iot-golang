package repository

import (
	"context"
	"fmt"
	"greenhouse-iot-golang/internal/model"

	"github.com/jmoiron/sqlx"
)

type sensorRepositoryImpl struct {
	db *sqlx.DB
}

func NewSensorRepository(db *sqlx.DB) SensorRepository {
	return &sensorRepositoryImpl{db: db}
}

func (r *sensorRepositoryImpl) Create(ctx context.Context, reading *model.SensorReading) error {
	query := `
		INSERT INTO sensor_readings (device_id, sensor_type, value, unit, recorded_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, public_id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		reading.DeviceID,
		reading.SensorType,
		reading.Value,
		reading.Unit,
		reading.RecordedAt,
	).Scan(&reading.ID, &reading.PublicID, &reading.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert sensor reading: %w", err)
	}

	return nil
}
