# parkiroid-server

Go API server for Parkiroid device frames and metrics.

## Docker

Build and run with Docker Compose:

```bash
docker compose up --build
```

The API listens on port `8080` by default. Override the host port with `PARKIROID_HOST_PORT`.

Set secrets for production:

```bash
export PARKIROID_API_KEY=your-api-key
export PARKIROID_JWT_SECRET=your-jwt-secret
docker compose up --build -d
```

SQLite and uploaded frame images are stored in the `parkiroid-data` Docker volume at `/data` inside the container.

### Docker only (without Compose)

```bash
docker build -t parkiroid-server .
docker run --rm -p 8080:8080 \
  -e PARKIROID_API_KEY=your-api-key \
  -e PARKIROID_JWT_SECRET=your-jwt-secret \
  -v parkiroid-data:/data \
  parkiroid-server
```

Health check: `GET http://localhost:8080/parkiroid/api/v1/health`
