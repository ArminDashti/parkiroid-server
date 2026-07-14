# API Endpoints

Base: `/dogan/api/v1`

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /auth | Login with username/password or Android api_key | No |
| POST | /auth/login | Web login (email/password) with user object | No |
| GET | /auth/me | Get current authenticated user | Yes |
| POST | /auth/logout | Log out current session | Yes |
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
| GET | /ai-models | List registered AI models (admin metadata) | Yes |
| POST | /ai-models | Register or update NCNN model metadata | Yes |
| GET | /models | List downloadable NCNN models for Android | Yes |
| GET | /models/:id/param | Download NCNN model.param file | Yes |
| GET | /models/:id/bin | Download NCNN model.bin file | Yes |
| GET | /webrtc/connections | List recent WebRTC sessions | Yes |
| POST | /streaming/token | Issue LiveKit WebRTC token | Yes |
| POST | /webrtc/session | Create LiveKit publisher session (Android) | Yes |
| GET | /devices | List registered devices | Yes |
| GET | /devices/:id/stream | Get LiveKit subscriber credentials (web) | Yes |
