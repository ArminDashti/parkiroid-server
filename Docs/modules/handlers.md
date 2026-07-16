# Handlers Module

Gin HTTP handlers for the Dogan REST API.

## Files
- `auth.go` — login, web login aliases, current user
- `frame.go` — upload/download frames
- `telemetry.go` — Android unified `POST /telemetry`
- `device_metrics.go` — legacy Android metrics
- `devices_web.go` — web device telemetry/metrics/capture/gallery
- `actions.go` — phone action queue
- `settings.go` — Android flat settings + web PATCH prefs
- `ai_models.go` — NCNN model registry, Android manifest, and file download
- `sounds.go` — alert sound manifest + file download
- `diagnostic_audio.go` — cabin diagnostic WAV upload
- `webrtc.go` — list WebRTC sessions
- `livekit.go` — issue LiveKit token + log connection
- `webrtc_session.go` — Android session alias + device stream credentials
- `health.go`, `endpoints.go` — health and self-describing API

## Auth
Protected routes accept JWT (from `/auth`) or embedded API token (`DOGAN_EMBEDDED_API_TOKEN`).
