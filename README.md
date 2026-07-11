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

1. Obtain a JWT from `POST /dogan/api/v1/auth`.
2. Request a LiveKit token from `POST /dogan/api/v1/streaming/token` with `device_id` and `role` (`publisher` for Android, `subscriber` for web).
3. Connect with a LiveKit client SDK using the returned `url`, `token`, and `room`.
