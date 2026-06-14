# parkiroid-server

Go API server for Parkiroid device frames, metrics, and LiveKit WebRTC streaming.

## Docker

Build and run with Docker Compose (includes LiveKit server):

```bash
docker compose up --build
```

The API listens on port `8080` by default. LiveKit listens on port `7880`. Override host ports with `PARKIROID_HOST_PORT` and `LIVEKIT_HOST_PORT`.

Set secrets for production:

```bash
export PARKIROID_JWT_SECRET=your-jwt-secret
export PARKIROID_LIVEKIT_API_KEY=your-livekit-api-key
export PARKIROID_LIVEKIT_API_SECRET=your-livekit-api-secret
export PARKIROID_LIVEKIT_URL=wss://your-livekit-host
docker compose up --build -d
```

Replace the bcrypt password hash in `internal/auth/credentials.go` before deploying.

SQLite and uploaded frame images are stored in the `parkiroid-data` Docker volume at `/data` inside the container.

### LiveKit streaming

1. Obtain a Parkiroid bearer token from `POST /parkiroid/api/v1/auth`.
2. Request a LiveKit token from `POST /parkiroid/api/v1/streaming/token` with `device_id` and `role` (`publisher` for devices, `subscriber` for viewers).
3. Connect with a LiveKit client SDK using the returned `url`, `token`, and `room`.

See `endpoints.md` for the full API contract.

### Docker only (without Compose)

```bash
docker build -t parkiroid-server .
docker run --rm -p 8080:8080 \
  -e PARKIROID_JWT_SECRET=your-jwt-secret \
  -v parkiroid-data:/data \
  parkiroid-server
```

Health check: `GET http://localhost:8080/parkiroid/api/v1/health`
