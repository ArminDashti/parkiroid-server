# Dogan API Endpoints

Base: `/dogan/api/v1`

| Endpoint | Description |
|----------|-------------|
| POST /auth | Login and get JWT token |
| GET /endpoints | List available API endpoints |
| GET /health | Health check |
| POST /frame | Upload camera frame from Android |
| GET /last-frame | Get latest frame metadata |
| GET /frame/image | Download latest frame JPEG |
| POST /device-metrics | Submit Android telemetry (battery, signal, network, temp, GPS) |
| GET /device-metrics | Get latest Android telemetry |
| POST /actions | Queue action for Android phone |
| GET /actions/pending | Poll pending actions for device |
| PUT /actions/:id/ack | Acknowledge action completion |
| GET /settings | Get app settings by platform |
| PUT /settings | Upsert app setting |
| GET /ai-models | List AI model download paths |
| POST /ai-models | Register or update AI model path |
| GET /webrtc/connections | List recent WebRTC sessions |
| POST /streaming/token | Issue LiveKit WebRTC token |
