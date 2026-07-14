# Dogan API Endpoints

Base: `/dogan/api/v1`

| Endpoint | Description |
|----------|-------------|
| POST /auth | Login with username/password or Android `api_key` |
| POST /auth/login | Web login (email/password) and get JWT token with user |
| GET /auth/me | Get current authenticated user |
| POST /auth/logout | Log out current session (204) |
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
| GET /ai-models | List registered AI models (admin metadata) |
| POST /ai-models | Register or update NCNN model metadata |
| GET /models | List downloadable NCNN models for Android |
| GET /models/:id/param | Download NCNN model.param file |
| GET /models/:id/bin | Download NCNN model.bin file |
| GET /webrtc/connections | List recent WebRTC sessions |
| POST /streaming/token | Issue LiveKit WebRTC token |
| POST /webrtc/session | Create LiveKit publisher session (Android) |
| GET /devices | List registered devices |
| GET /devices/:id/stream | Get LiveKit subscriber credentials for web viewer |
