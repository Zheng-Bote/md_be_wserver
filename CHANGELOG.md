# Changelog

All notable changes to the `md_be_wserver` project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-08-17

### Added
- **Medical Devices API:** New endpoints `/api/v1/devices/get_device_types` and `/api/v1/devices/get_devices` for retrieving device configurations.
- **Google Flatbuffers Integration:** Implemented high-performance binary serialization using Flatbuffers (`fbs/devices.fbs`) for device data exchange.
- **Database Schema:** Added `devices` and `device_types` tables along with seed data (`migrations/02_devices_schema.sql`).

### Fixed
- **Postgres Error Handling:** Fixed a bug where `rows.Err()` was not checked in `pgx` queries, which previously caused `permission denied` database errors to be silently swallowed instead of returning an `HTTP 500`.

## [0.1.0] - 2026-08-16

### Added
- Initial project structure for the Medical Devices Backend Webserver (`md_be_wserver`).
- PostgreSQL v18.4+ database schema design (`migrations/schema.sql`) including Users, Profiles, Roles, Permissions, and Sessions.
- Database connection handling and robust connection pooling using `pgxpool`.
- Automatic initialization and synchronisation of admin users from JSON config to the database upon server startup.
- Full Auth-Flow with JSON Web Tokens (JWT) for stateless authorization.
- Token-based `/api/v1/auth/login` and `/api/v1/auth/refresh` endpoints.
- Argon2Id password hashing for highly secure credential storage.
- Forced password change mechanism for first-time logins (setting `must_change_pwd`).
- Endpoint `/api/v1/auth/change-password` for user password management.
- Complete TOTP (Time-Based One-Time Password) infrastructure for 2FA.
- Endpoints `/api/v1/auth/totp/setup` (including PNG Base64 QR-Code generation) and `/api/v1/auth/totp/verify`.
- Robust JWT Middleware (`RequireAuth`) for route protection and strict 2FA/Password-Change enforcement.
- Configuration structure supporting AES-256-GCM encrypted configurations (`wserver` object).
- Custom `encrypt-config` CLI tool to encrypt the configuration file with a password.
- Standardized File Headers (SPDX & Doxygen) across all Go source files.
- Standalone `NOTICE` file outlining all used third-party Open Source dependencies.
