/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file handlers_devices.go
 * @brief HTTP handlers for devices API
 * @version 1.0.0
 * @date 2026-08-17
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package http

import (
	"net/http"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"md_be_wserver/internal/fbs"
)

func (s *Server) handleGetDeviceTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	deviceTypes, err := s.Repo.GetDeviceTypes(ctx)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	builder := flatbuffers.NewBuilder(1024)
	
	var dtOffsets []flatbuffers.UOffsetT
	for _, dt := range deviceTypes {
		idOffset := builder.CreateString(dt.ID)
		nameOffset := builder.CreateString(dt.Name)
		descOffset := builder.CreateString(dt.Description)
		createdAtOffset := builder.CreateString(dt.CreatedAt.Format(time.RFC3339))
		lastUpdateOffset := builder.CreateString(dt.LastUpdate.Format(time.RFC3339))

		fbs.DeviceTypeStart(builder)
		fbs.DeviceTypeAddId(builder, idOffset)
		fbs.DeviceTypeAddName(builder, nameOffset)
		fbs.DeviceTypeAddDescription(builder, descOffset)
		fbs.DeviceTypeAddCreatedAt(builder, createdAtOffset)
		fbs.DeviceTypeAddLastUpdate(builder, lastUpdateOffset)
		fbs.DeviceTypeAddActive(builder, dt.Active)
		dtOffsets = append(dtOffsets, fbs.DeviceTypeEnd(builder))
	}

	fbs.DeviceTypeListStartTypesVector(builder, len(dtOffsets))
	for i := len(dtOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(dtOffsets[i])
	}
	typesVector := builder.EndVector(len(dtOffsets))

	fbs.DeviceTypeListStart(builder)
	fbs.DeviceTypeListAddTypes(builder, typesVector)
	root := fbs.DeviceTypeListEnd(builder)
	builder.Finish(root)

	w.Header().Set("Content-Type", "application/x-flatbuffers")
	w.Write(builder.FinishedBytes())
}

func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx := r.Context()
	
	typeID := r.URL.Query().Get("device_type")
	deviceID := r.URL.Query().Get("id")

	devices, err := s.Repo.GetDevices(ctx, typeID, deviceID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	builder := flatbuffers.NewBuilder(1024)
	
	var devOffsets []flatbuffers.UOffsetT
	for _, d := range devices {
		idOffset := builder.CreateString(d.ID)
		typeIdOffset := builder.CreateString(d.TypeID)
		devNameOffset := builder.CreateString(d.DeviceName)
		manufOffset := builder.CreateString(d.Manufacturer)
		ifaceOffset := builder.CreateString(d.Interface)
		descOffset := builder.CreateString(d.Description)
		createdAtOffset := builder.CreateString(d.CreatedAt.Format(time.RFC3339))
		lastUpdateOffset := builder.CreateString(d.LastUpdate.Format(time.RFC3339))

		fbs.DeviceStart(builder)
		fbs.DeviceAddId(builder, idOffset)
		fbs.DeviceAddTypeId(builder, typeIdOffset)
		fbs.DeviceAddDeviceName(builder, devNameOffset)
		fbs.DeviceAddManufacturer(builder, manufOffset)
		fbs.DeviceAddInterface(builder, ifaceOffset)
		fbs.DeviceAddDescription(builder, descOffset)
		fbs.DeviceAddCreatedAt(builder, createdAtOffset)
		fbs.DeviceAddLastUpdate(builder, lastUpdateOffset)
		fbs.DeviceAddActive(builder, d.Active)
		devOffsets = append(devOffsets, fbs.DeviceEnd(builder))
	}

	fbs.DeviceListStartDevicesVector(builder, len(devOffsets))
	for i := len(devOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(devOffsets[i])
	}
	devicesVector := builder.EndVector(len(devOffsets))

	fbs.DeviceListStart(builder)
	fbs.DeviceListAddDevices(builder, devicesVector)
	root := fbs.DeviceListEnd(builder)
	builder.Finish(root)

	w.Header().Set("Content-Type", "application/x-flatbuffers")
	w.Write(builder.FinishedBytes())
}
