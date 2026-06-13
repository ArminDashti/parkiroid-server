# Parkiroid Server API Endpoints

Base path: `/parkiroid/api/v1`

Default listen address: `:8080` (configurable via `PARKIROID_LISTEN_ADDRESS`)

---

## Authentication

Public endpoints do not require a token.

Protected endpoints require a bearer token obtained from `POST /auth`.

```
Authorization: Bearer <token>
```

Obtain a token by exchanging the API key (`PARKIROID_API_KEY`, default: `parkiroid-dev-key`).

---

## Endpoints

### POST `/parkiroid/api/v1/auth`

Exchange API key for a bearer token.

**Auth required:** No

**Request body:**

```json
{
  "api_key": "parkiroid-dev-key"
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
| `401` | `invalid api key` |
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
      "description": "Exchange API key for a bearer token",
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

## Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/parkiroid/api/v1/auth` | No | Exchange API key for a bearer token |
| `GET` | `/parkiroid/api/v1/endpoints` | No | List available API endpoints |
| `GET` | `/parkiroid/api/v1/health` | No | Service health check |
| `POST` | `/parkiroid/api/v1/frame` | Yes | Submit a camera frame from a device |
| `GET` | `/parkiroid/api/v1/last-frame` | Yes | Retrieve the most recent frame for a device |
| `POST` | `/parkiroid/api/v1/device-metrics` | Yes | Submit device telemetry metrics |
| `GET` | `/parkiroid/api/v1/device-metrics` | Yes | Retrieve the latest metrics for a device |
