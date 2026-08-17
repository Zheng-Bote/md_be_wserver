/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file devices.go
 * @brief Database operations for devices
 * @version 1.0.0
 * @date 2026-08-17
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package db

import (
	"context"
	"fmt"
	"time"
)

type DeviceType struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	LastUpdate  time.Time
	Active      bool
}

type Device struct {
	ID           string
	TypeID       string
	DeviceName   string
	Manufacturer string
	Interface    string
	Description  string
	CreatedAt    time.Time
	LastUpdate   time.Time
	Active       bool
}

func (r *Repository) GetDeviceTypes(ctx context.Context) ([]DeviceType, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name, description, created_at, last_update, active FROM device_types`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []DeviceType
	for rows.Next() {
		var dt DeviceType
		var desc *string
		if err := rows.Scan(&dt.ID, &dt.Name, &desc, &dt.CreatedAt, &dt.LastUpdate, &dt.Active); err != nil {
			return nil, err
		}
		if desc != nil {
			dt.Description = *desc
		}
		types = append(types, dt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return types, nil
}

func (r *Repository) GetDevices(ctx context.Context, typeID, deviceID string) ([]Device, error) {
	query := `SELECT id, type_id, device_name, manufacturer, interface, description, created_at, last_update, active FROM devices WHERE 1=1`
	var args []interface{}
	argCount := 1

	if typeID != "" {
		query += fmt.Sprintf(" AND type_id = $%d", argCount)
		args = append(args, typeID)
		argCount++
	}
	if deviceID != "" {
		query += fmt.Sprintf(" AND id = $%d", argCount)
		args = append(args, deviceID)
		argCount++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		var tId, manuf, iface, desc *string
		if err := rows.Scan(&d.ID, &tId, &d.DeviceName, &manuf, &iface, &desc, &d.CreatedAt, &d.LastUpdate, &d.Active); err != nil {
			return nil, err
		}
		if tId != nil {
			d.TypeID = *tId
		}
		if manuf != nil {
			d.Manufacturer = *manuf
		}
		if iface != nil {
			d.Interface = *iface
		}
		if desc != nil {
			d.Description = *desc
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return devices, nil
}
