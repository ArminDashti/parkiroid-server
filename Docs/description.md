# Dogan Server

Dogan is a Go REST API for a personal Android + web monitoring stack. It stores camera frames, Android telemetry, app settings, NCNN AI model metadata and binaries, and phone actions in PostgreSQL, and issues LiveKit tokens for WebRTC streaming between Android and the web client.

**Stack:** Go 1.26, Gin, PostgreSQL 16, LiveKit, Docker Compose.

**Run:**
```powershell
.\build-docker-image.ps1
docker compose up -d
.\scripts\test-dummy-data.ps1
```

**Entry point:** `cmd/server/main.go`  
**API base:** `/dogan/api/v1`  
**Login:** user `armin` — see `armin-credentials.txt` (local, gitignored)
