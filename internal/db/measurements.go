package db

import (
	"context"
)

type Measurement struct {
	ID               string `json:"measurement_id"`
	DeviceID         string `json:"device_id"`
	PayloadEncrypted string `json:"payload_encrypted"`
}

func (r *Repository) SaveMeasurements(ctx context.Context, measurements []Measurement) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Note: We removed the CREATE TABLE IF NOT EXISTS here because the DB user
    // lacks DDL permissions on the public schema. The table must be created
    // via a migration script (e.g. migrations/03_measurements_schema.sql)

	for _, m := range measurements {
		_, err := tx.Exec(ctx, `
			INSERT INTO measurements (id, device_id, payload_encrypted)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, m.ID, m.DeviceID, m.PayloadEncrypted)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
