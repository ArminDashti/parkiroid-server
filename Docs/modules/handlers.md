# Handlers Module

Gin HTTP handlers for the Dogan REST API.

## Files
- `auth.go` — login + login_logs audit
- `frame.go` — upload/download frames
- `device_metrics.go` — Android telemetry
- `actions.go` — phone action queue
- `settings.go` — web/android settings
- `ai_models.go` — AI model path registry
- `webrtc.go` — list WebRTC sessions
- `livekit.go` — issue LiveKit token + log connection
- `health.go`, `endpoints.go` — health and self-describing API

## Auth
Protected routes accept JWT (from `/auth`) or embedded API token (`DOGAN_EMBEDDED_API_TOKEN`).
