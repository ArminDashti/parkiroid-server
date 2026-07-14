# dogan-server

Go API server for Dogan Android/Web clients: REST image and telemetry transfer, PostgreSQL persistence, and LiveKit WebRTC streaming.

## Docker

Build and run with Docker Compose (includes PostgreSQL and LiveKit):

```powershell
.\build-docker-image.ps1
docker compose up -d
```

The API listens on port `8080` by default. PostgreSQL on `5432`. LiveKit on `7880`.

Login credentials for the web app are in `armin-credentials.txt` (generated locally, gitignored).

Generate an embedded API token for Android client apps:

```bash
go run ./cmd/issue-token
```

Set the same value on the server as `DOGAN_EMBEDDED_API_TOKEN`.

## API

See `endpoints.md` for the endpoint list.

Base path: `/dogan/api/v1`

Health check: `GET http://localhost:8080/dogan/api/v1/health`

Test with dummy data after Docker is running:

```powershell
.\scripts\test-dummy-data.ps1
```

## WebRTC

1. Obtain a JWT from `POST /dogan/api/v1/auth` (or `POST /auth/login` from the web app).
2. **Android:** call `POST /dogan/api/v1/webrtc/session` with `device_id` (or `POST /streaming/token` with `role: publisher`).
3. **Web:** call `GET /dogan/api/v1/devices/:id/stream` for subscriber credentials.
4. Connect with a LiveKit client SDK using the returned `url`, `token`, and `room`.

Set `DOGAN_LIVEKIT_PUBLIC_URL` (e.g. `ws://localhost:7880`) so clients receive a reachable LiveKit URL when the API runs in Docker.

## AI models (NCNN)

Place NCNN files on the server under `DOGAN_MODELS_DIR` (Docker default `/data/models`):

```
{model_id}/model.param
{model_id}/model.bin
```

Model ids must match the Android app (`yolov8_nano`, `yolov8_small`, `mobilenet_ssd`). Register metadata:

```powershell
.\scripts\register-models.ps1
```

Android fetches `GET /dogan/api/v1/models` and downloads `.param`/`.bin` with bearer auth.
