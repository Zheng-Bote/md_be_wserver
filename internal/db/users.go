/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file users.go
 * @brief User management and authentication operations
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

	"md_be_wserver/internal/config"
	"md_be_wserver/internal/crypto"
)

// SyncAdmins checks if the configured admin users exist in the database,
// and creates them with 'must_change_pwd' = true if they do not exist.
func (r *Repository) SyncAdmins(ctx context.Context, admins []config.AdminUser) error {
	for _, admin := range admins {
		var exists bool
		err := r.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE login_id = $1)", admin.Username).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check user %s: %w", admin.Username, err)
		}

		if !exists {
			log.Printf("Admin user '%s' not found in DB. Creating...", admin.Username)

			hashedPassword, err := crypto.HashPassword(admin.Password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", admin.Username, err)
			}

			_, err = r.Pool.Exec(ctx, `
				INSERT INTO users (login_id, password_hash, is_active, must_change_pwd) 
				VALUES ($1, $2, true, true)`,
				admin.Username, hashedPassword)
			if err != nil {
				return fmt.Errorf("failed to insert user %s: %w", admin.Username, err)
			}
			log.Printf("Successfully created admin user '%s' (Must change password on first login)", admin.Username)
		}
	}
	return nil
}

type User struct {
	ID                  string
	LoginID             string
	PasswordHash        string
	IsActive            bool
	MustChangePwd       bool
	FailedLoginAttempts int
	LockedUntil         *time.Time
	TotpEnabled         bool
	TotpSecret          *string
}

func (r *Repository) GetUserByLoginID(ctx context.Context, loginID string) (*User, error) {
	var u User
	err := r.Pool.QueryRow(ctx, `
		SELECT id, login_id, password_hash, is_active, must_change_pwd, failed_login_attempts, locked_until, totp_enabled, totp_secret 
		FROM users WHERE login_id = $1`, loginID).Scan(
		&u.ID, &u.LoginID, &u.PasswordHash, &u.IsActive, &u.MustChangePwd, &u.FailedLoginAttempts, &u.LockedUntil, &u.TotpEnabled, &u.TotpSecret)
	return &u, err
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.Pool.QueryRow(ctx, `
		SELECT id, login_id, password_hash, is_active, must_change_pwd, failed_login_attempts, locked_until, totp_enabled, totp_secret 
		FROM users WHERE id = $1`, id).Scan(
		&u.ID, &u.LoginID, &u.PasswordHash, &u.IsActive, &u.MustChangePwd, &u.FailedLoginAttempts, &u.LockedUntil, &u.TotpEnabled, &u.TotpSecret)
	return &u, err
}

func (r *Repository) UpdateTOTPSecret(ctx context.Context, id string, secret string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE users SET totp_secret = $1 WHERE id = $2`, secret, id)
	return err
}

func (r *Repository) EnableTOTP(ctx context.Context, id string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE users SET totp_enabled = true WHERE id = $1`, id)
	return err
}

func (r *Repository) UpdateFailedLogins(ctx context.Context, id string, attempts int, lockedUntil *time.Time) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE users SET failed_login_attempts = $1, locked_until = $2 WHERE id = $3`,
		attempts, lockedUntil, id)
	return err
}

func (r *Repository) ResetFailedLogins(ctx context.Context, id string) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = $1`, id)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, id string, newHash string) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, must_change_pwd = false, last_pwd_change = CURRENT_TIMESTAMP WHERE id = $2`,
		newHash, id)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, userID, ip, userAgent string, expiresAt time.Time) (string, error) {
	var sessionID string
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, ip, userAgent, expiresAt).Scan(&sessionID)
	return sessionID, err
}

func (r *Repository) GetSession(ctx context.Context, sessionID string) (string, error) {
	var userID string
	err := r.Pool.QueryRow(ctx, `
		SELECT user_id FROM sessions WHERE id = $1 AND is_revoked = false AND expires_at > CURRENT_TIMESTAMP`,
		sessionID).Scan(&userID)
	return userID, err
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE sessions SET is_revoked = true WHERE id = $1`, sessionID)
	return err
}
