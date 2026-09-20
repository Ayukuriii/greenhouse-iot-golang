package repository

import (
	"context"
	"fmt"
	"greenhouse-iot-golang/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type deviceRepositoryImpl struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
	return &deviceRepositoryImpl{db: db}
}

func (r *deviceRepositoryImpl) CreateCommand(ctx context.Context, cmd *model.DeviceCommand) error {
	query := `
		INSERT INTO device_commands (device_id, command, status)
		VALUES ($1, $2, $3)
		RETURNING id, public_id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		cmd.DeviceID,
		cmd.Command,
		cmd.Status,
	).Scan(&cmd.ID, &cmd.PublicID, &cmd.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create device command: %w", err)
	}

	return nil
}

func (r *deviceRepositoryImpl) UpdateCommandStatus(ctx context.Context, id int64, status string, publishedAt *time.Time) error {
	query := `
		UPDATE device_commands
		SET status = $1, published_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, publishedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update device command status: %w", err)
	}
	return nil
}
