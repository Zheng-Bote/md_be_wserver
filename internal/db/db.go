/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file db.go
 * @brief PostgreSQL database connection and repository setup
 * @version 1.0.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"md_be_wserver/internal/config"
)

type Repository struct {
	Pool *pgxpool.Pool
}

func NewRepository(cfg config.DBConnectionConfig) (*Repository, error) {
	sslMode := "disable"
	if cfg.SSLMode {
		sslMode = "require"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, sslMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Wait before connecting if db_connect_delay is set
	if cfg.DBConnectDelay > 0 {
		log.Printf("Waiting %d seconds before connecting to DB...", cfg.DBConnectDelay)
		time.Sleep(time.Duration(cfg.DBConnectDelay) * time.Second)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database to ensure connection is valid
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Repository{Pool: pool}, nil
}

func (r *Repository) Close() {
	if r.Pool != nil {
		r.Pool.Close()
	}
}
