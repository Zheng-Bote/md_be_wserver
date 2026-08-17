# API Endpoints Overview

| Endpoint                       | Method | Auth Required | Description                                         |
| ------------------------------ | ------ | ------------- | --------------------------------------------------- |
| `/api/v1/auth/login`           | POST   | No            | Authenticate user, returns JWT and Refresh Token.   |
| `/api/v1/auth/refresh`         | POST   | No            | Issue new JWT using valid Refresh Token.            |
| `/api/v1/auth/logout`          | POST   | Yes           | Revokes the current session refresh token.          |
| `/api/v1/auth/change-password` | POST   | Yes           | Updates user password (enforced on first login).    |
| `/api/v1/auth/totp/setup`      | POST   | Yes           | Generates 2FA Secret and Base64 PNG QR Code.        |
| `/api/v1/auth/totp/verify`     | POST   | Yes           | Verifies initial code to enable 2FA on the account. |
| `/api/v1/devices/get_device_types` | GET | Yes | Returns a list of all available device types (FlatBuffers). |
| `/api/v1/devices/get_devices` | GET | Yes | Returns all devices, or filters by `?device_type=<UUID>` or `?id=<UUID>` (FlatBuffers). |
