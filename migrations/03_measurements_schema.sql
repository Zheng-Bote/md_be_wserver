-- SPDX-FileComment: Medical Devices Backend
-- SPDX-FileType: SOURCE
-- SPDX-FileContributor: ZHENG Robert
-- SPDX-FileCopyrightText: 2026 ZHENG Robert
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS measurements (
    id UUID PRIMARY KEY,
    device_id UUID NOT NULL,
    payload_encrypted TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Give the mitm_user permissions to write to it
GRANT SELECT, INSERT, UPDATE, DELETE ON measurements TO mitm_user;
