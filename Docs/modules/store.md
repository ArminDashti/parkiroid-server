# Store Module

PostgreSQL persistence for all Dogan data.

## Tables
- `login_logs` — login audit (IP, user, browser, success/fail)
- `stored_images` — JPEG file paths per device
- `phone_actions` — commands queued for Android
- `webrtc_connections` — LiveKit session log
- `app_settings` — key/value settings per platform (web/android)
- `ai_model_paths` — downloadable AI model paths
- `android_telemetry` — battery, signal, network, temp, GPS
- `devices` — device registry helper

## Key files
- `schema.go` — DDL applied on startup
- `postgres.go` — CRUD implementations
- `retention.go` — background cleanup of old frames/metrics
- `frame_storage.go` — writes JPEG files to disk

## Config
`DOGAN_DATABASE_URL` (default `postgres://dogan:dogan@postgres:5432/dogan?sslmode=disable`)
