# API Endpoints

Base: `/dogan/api/v1`

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /auth | Login and get JWT token | No |
| GET | /endpoints | List available API endpoints | No |
| GET | /health | Health check | No |
| POST | /frame | Upload camera frame from Android | Yes |
| GET | /last-frame | Get latest frame metadata | Yes |
| GET | /frame/image | Download latest frame JPEG | Yes |
| POST | /device-metrics | Submit Android telemetry | Yes |
| GET | /device-metrics | Get latest Android telemetry | Yes |
| POST | /actions | Queue action for Android phone | Yes |
| GET | /actions/pending | Poll pending actions for device | Yes |
| PUT | /actions/:id/ack | Acknowledge action completion | Yes |
| GET | /settings | Get app settings by platform | Yes |
| PUT | /settings | Upsert app setting | Yes |
| GET | /ai-models | List AI model download paths | Yes |
| POST | /ai-models | Register or update AI model path | Yes |
| GET | /webrtc/connections | List recent WebRTC sessions | Yes |
| POST | /streaming/token | Issue LiveKit WebRTC token | Yes |
