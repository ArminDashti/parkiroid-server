# Store Module

PostgreSQL persistence for all Dogan data.

## Tables
- `login_logs` — login audit (IP, user, browser, success/fail)
- `stored_images` — JPEG file paths per device
- `phone_actions` — commands queued for Android
- `webrtc_connections` — LiveKit session log
- `app_settings` — key/value settings per platform (web/android)
- `ai_model_paths` — NCNN model metadata (SHA-256, labels, format)
- `android_telemetry` — battery, signal, network, temp, GPS
- `devices` — device registry helper

## Key files
- `schema.go` — DDL applied on startup
- `postgres.go` — CRUD implementations (includes `ListDevices` for web device picker)
- `retention.go` — background cleanup of old frames/metrics
- `model_files.go` — NCNN file paths, SHA-256, on-disk presence checks

## Config
`DOGAN_DATABASE_URL` (default `postgres://dogan:dogan@postgres:5432/dogan?sslmode=disable`)  
`DOGAN_MODELS_DIR` (default `models`, Docker `/data/models`) — on-disk layout `{id}/model.param` and `{id}/model.bin`
