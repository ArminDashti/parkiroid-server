# Handlers Module

Gin HTTP handlers for the Dogan REST API.

## Files
- `auth.go` — login, web login aliases, current user
- `frame.go` — upload/download frames
- `device_metrics.go` — Android telemetry
- `actions.go` — phone action queue
- `settings.go` — web/android settings
- `ai_models.go` — NCNN model registry, Android manifest, and file download
- `webrtc.go` — list WebRTC sessions
- `livekit.go` — issue LiveKit token + log connection
- `webrtc_session.go` — Android session alias + device stream credentials
- `devices.go` — list devices for web UI
- `health.go`, `endpoints.go` — health and self-describing API

## Auth
Protected routes accept JWT (from `/auth`) or embedded API token (`DOGAN_EMBEDDED_API_TOKEN`).
