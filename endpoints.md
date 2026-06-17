# Parkiroid Server API Endpoints

Base path: `/parkiroid/api/v1`

Default listen address: `:8080` (configurable via `PARKIROID_LISTEN_ADDRESS`)

---

## Authentication

Public endpoints do not require a token.

Protected endpoints require a bearer token.

```
Authorization: Bearer <token>
```

Two token types are accepted:

1. **Embedded API token** (recommended for client apps) — a static token configured on the server via `PARKIROID_EMBEDDED_API_TOKEN`. Embed the same value in your app. Development default:

   ```
   pk_dev_a8f3c2e1b9d74f6a0e5c3b9d2f7a1e4c8b6d0f3a7e2c9b5d1f8a4e6c0b3d7f9
   ```

   Generate a production token with:

   ```bash
   go run ./cmd/issue-token
   ```

2. **JWT from `POST /auth`** — short-lived admin token obtained by posting credentials. The server verifies a single hardcoded account (`admin` / `parkiroid-dev-password` in development). Replace the bcrypt hash in `internal/auth/credentials.go` for production.

---

## Endpoints

### POST `/parkiroid/api/v1/auth`

Exchange admin credentials for a bearer token.

**Auth required:** No

**Request body:**

```json
{
  "username": "admin",
  "password": "parkiroid-dev-password"
}
```

**Success response:** `200 OK`

```json
{
  "token": "<jwt>",
  "expires_at": "2026-06-10T12:00:00Z"
}
```

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `invalid request body` |
| `401` | `invalid username or password` |
| `500` | `failed to issue token` |

---

### GET `/parkiroid/api/v1/endpoints`

List all available API endpoints.

**Auth required:** No

**Success response:** `200 OK`

```json
{
  "endpoints": [
    {
      "method": "POST",
      "path": "/parkiroid/api/v1/auth",
      "description": "Exchange admin credentials for a bearer token",
      "auth_required": false
    },
    {
      "method": "GET",
      "path": "/parkiroid/api/v1/endpoints",
      "description": "List available API endpoints",
      "auth_required": false
    },
    {
      "method": "GET",
      "path": "/parkiroid/api/v1/health",
      "description": "Service health check",
      "auth_required": false
    },
    {
      "method": "GET",
      "path": "/parkiroid/api/v1/last-frame",
      "description": "Retrieve the most recent frame for a device",
      "auth_required": true
    },
    {
      "method": "POST",
      "path": "/parkiroid/api/v1/frame",
      "description": "Submit a camera frame from a device",
      "auth_required": true
    },
    {
      "method": "GET",
      "path": "/parkiroid/api/v1/device-metrics",
      "description": "Retrieve the latest metrics for a device",
      "auth_required": true
    },
    {
      "method": "POST",
      "path": "/parkiroid/api/v1/device-metrics",
      "description": "Submit device telemetry metrics",
      "auth_required": true
    },
    {
      "method": "POST",
      "path": "/parkiroid/api/v1/streaming/token",
      "description": "Issue a LiveKit access token for WebRTC streaming",
      "auth_required": true
    }
  ]
}
```

---

### GET `/parkiroid/api/v1/health`

Service health check.

**Auth required:** No

**Success response:** `200 OK`

```json
{
  "status": "ok",
  "timestamp": "2026-06-13T12:00:00Z"
}
```

---

### POST `/parkiroid/api/v1/frame`

Submit a camera frame from a device.

**Auth required:** Yes

**Request body:**

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "image_data": "<base64-encoded-jpeg>",
  "captured_at": "2026-06-13T12:00:00Z"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `device_id` | Yes | Device identifier (MAC address, device name, or numeric ID) |
| `image_data` | Yes | Base64-encoded JPEG. Data-URI prefixes (`data:image/jpeg;base64,...`) are accepted |
| `captured_at` | No | Capture timestamp (RFC 3339). Defaults to server time if omitted |

**Success response:** `201 Created`

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "path": "frames/aa-bb-cc-dd-ee-ff-060613-120000.jpg",
  "captured_at": "2026-06-13T12:00:00Z",
  "received_at": "2026-06-13T12:00:01Z"
}
```

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `invalid request body` |
| `400` | `invalid image_data` |
| `401` | `missing authorization header` / `invalid authorization header format` / `invalid or expired token` |
| `500` | `failed to save frame` |

---

### GET `/parkiroid/api/v1/last-frame`

Retrieve the most recent frame for a device.

**Auth required:** Yes

**Query parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `device-id` | Yes | Device identifier (MAC address, device name, or numeric ID) |

**Example:** `GET /parkiroid/api/v1/last-frame?device-id=aa:bb:cc:dd:ee:ff`

**Success response:** `200 OK`

```json
{
  "id": 1,
  "device_id": "aa:bb:cc:dd:ee:ff",
  "path": "frames/aa-bb-cc-dd-ee-ff-060613-120000.jpg",
  "captured_at": "2026-06-13T12:00:00Z"
}
```

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `device-id query parameter is required` |
| `401` | `missing authorization header` / `invalid authorization header format` / `invalid or expired token` |
| `404` | `no frame found for device` |
| `500` | `failed to retrieve frame` |

---

### POST `/parkiroid/api/v1/device-metrics`

Submit device telemetry metrics.

**Auth required:** Yes

**Request body:**

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "cpu_usage_percent": 42.5,
  "memory_usage_percent": 68.0,
  "disk_usage_percent": 55.0,
  "battery_level_percent": 87.0,
  "temperature_celsius": 38.2,
  "signal_strength_dbm": -65,
  "recorded_at": "2026-06-13T12:00:00Z"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `device_id` | Yes | Device identifier |
| `cpu_usage_percent` | No | CPU usage percentage |
| `memory_usage_percent` | No | Memory usage percentage |
| `disk_usage_percent` | No | Disk usage percentage |
| `battery_level_percent` | No | Battery level percentage |
| `temperature_celsius` | No | Device temperature in Celsius |
| `signal_strength_dbm` | No | Signal strength in dBm |
| `recorded_at` | No | Metric timestamp (RFC 3339). Defaults to server time if omitted |

**Success response:** `201 Created`

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "cpu_usage_percent": 42.5,
  "memory_usage_percent": 68.0,
  "disk_usage_percent": 55.0,
  "battery_level_percent": 87.0,
  "temperature_celsius": 38.2,
  "signal_strength_dbm": -65,
  "recorded_at": "2026-06-13T12:00:00Z",
  "received_at": "2026-06-13T12:00:01Z"
}
```

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `invalid request body` |
| `401` | `missing authorization header` / `invalid authorization header format` / `invalid or expired token` |

---

### GET `/parkiroid/api/v1/device-metrics`

Retrieve the latest metrics for a device.

**Auth required:** Yes

**Query parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `device-id` | Yes | Device identifier |

**Example:** `GET /parkiroid/api/v1/device-metrics?device-id=aa:bb:cc:dd:ee:ff`

**Success response:** `200 OK`

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "cpu_usage_percent": 42.5,
  "memory_usage_percent": 68.0,
  "disk_usage_percent": 55.0,
  "battery_level_percent": 87.0,
  "temperature_celsius": 38.2,
  "signal_strength_dbm": -65,
  "recorded_at": "2026-06-13T12:00:00Z",
  "received_at": "2026-06-13T12:00:01Z"
}
```

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `device-id query parameter is required` |
| `401` | `missing authorization header` / `invalid authorization header format` / `invalid or expired token` |
| `404` | `no metrics found for device` |
| `500` | `failed to retrieve metrics` |

---

### POST `/parkiroid/api/v1/streaming/token`

Issue a LiveKit access token for WebRTC streaming to or from a device room.

Each device maps to a LiveKit room named `device-{sanitized-device-id}`. Devices publish video with role `publisher`; viewers subscribe with role `subscriber`.

**Auth required:** Yes

**Request body:**

```json
{
  "device_id": "aa:bb:cc:dd:ee:ff",
  "identity": "viewer-1",
  "role": "subscriber"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `device_id` | Yes | Device identifier (MAC address, device name, or numeric ID) |
| `identity` | No | LiveKit participant identity. Defaults to `publisher-{device-id}` or `subscriber-{device-id}` based on role |
| `role` | No | `publisher` (can publish tracks) or `subscriber` (subscribe only). Defaults to `subscriber` |

**Success response:** `200 OK`

```json
{
  "token": "<livekit-jwt>",
  "url": "ws://localhost:7880",
  "room": "device-aa-bb-cc-dd-ee-ff",
  "identity": "subscriber-aa-bb-cc-dd-ee-ff",
  "expires_at": "2026-06-13T13:00:00Z"
}
```

Use `url`, `token`, and `room` with a LiveKit client SDK to connect.

**Error responses:**

| Status | Error |
|--------|-------|
| `400` | `invalid request body` |
| `400` | `role must be publisher or subscriber` |
| `401` | `missing authorization header` / `invalid authorization header format` / `invalid or expired token` |
| `503` | `livekit is not configured` |
| `500` | `failed to issue livekit token` |

**Environment variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `PARKIROID_LIVEKIT_URL` | `ws://localhost:7880` | WebSocket URL returned to clients |
| `PARKIROID_LIVEKIT_API_KEY` | `devkey` | LiveKit API key |
| `PARKIROID_LIVEKIT_API_SECRET` | `secret` | LiveKit API secret |
| `PARKIROID_LIVEKIT_TOKEN_TTL` | `3600` | LiveKit token lifetime in seconds |

---

## Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/parkiroid/api/v1/auth` | No | Exchange admin credentials for a bearer token |
| `GET` | `/parkiroid/api/v1/endpoints` | No | List available API endpoints |
| `GET` | `/parkiroid/api/v1/health` | No | Service health check |
| `POST` | `/parkiroid/api/v1/frame` | Yes | Submit a camera frame from a device |
| `GET` | `/parkiroid/api/v1/last-frame` | Yes | Retrieve the most recent frame for a device |
| `POST` | `/parkiroid/api/v1/device-metrics` | Yes | Submit device telemetry metrics |
| `GET` | `/parkiroid/api/v1/device-metrics` | Yes | Retrieve the latest metrics for a device |
| `POST` | `/parkiroid/api/v1/streaming/token` | Yes | Issue a LiveKit access token for WebRTC streaming |
