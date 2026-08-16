/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file config.go
 * @brief Configuration data types and config loading
 * @version 0.1.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"md_be_wserver/internal/crypto"
)

type AdminUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DBConnectionConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Database       string `json:"database"`
	DBConnectDelay int    `json:"db_connect_delay,omitempty"`
	SSLMode        bool   `json:"sslmode,omitempty"`
}

type WebServerConfig struct {
	LogLevel          string `json:"log_level"`
	UploadDir         string `json:"upload_dir"`
	HTTPPort          int    `json:"http_port,omitempty"`
	UseHTTPS          bool   `json:"use_https,omitempty"`
	SSLCert           string `json:"ssl_cert,omitempty"`
	SSLKey            string `json:"ssl_key,omitempty"`
	RequireTOTP       bool   `json:"require_totp,omitempty"`
	JWTSecret         string `json:"jwt_secret"`
	ReadTimeout       int    `json:"read_timeout"`
	ReadHeaderTimeout int    `json:"read_header_timeout"`
	WriteTimeout      int    `json:"write_timeout"`
	IdleTimeout       int    `json:"idle_timeout"`
	LimitBody         int64  `json:"limit_body"`
	Throttle          int    `json:"throttle"`
}

type DBConfig struct {
	DB        DBConnectionConfig `json:"db"`
	Admins    []AdminUser        `json:"admins"`
	WServer   WebServerConfig    `json:"wserver"`
}

// LoadConfig reads an encrypted or plaintext JSON file into DBConfig
func LoadConfig(filePath string, password string) (*DBConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var dbConfig DBConfig

	// Try to unmarshal as plaintext first
	if err := json.Unmarshal(data, &dbConfig); err == nil {
		return &dbConfig, nil
	}

	// If it fails, assume it's encrypted
	decryptedData, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config (also not valid plaintext JSON): %w", err)
	}

	if err := json.Unmarshal(decryptedData, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted config JSON: %w", err)
	}

	return &dbConfig, nil
}

func (c *DBConfig) GetDSN() string {
	sslMode := "disable"
	if c.DB.SSLMode {
		sslMode = "require"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Database, sslMode)
}
