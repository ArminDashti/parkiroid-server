# Dogan Server

Dogan is a Go REST API for a personal Android + web monitoring stack. It stores camera frames, Android telemetry, app settings, NCNN AI model metadata and binaries, and phone actions in PostgreSQL, and issues LiveKit tokens for WebRTC streaming between Android and the web client.

**Stack:** Go 1.26, Gin, PostgreSQL 16, LiveKit, Docker Compose.

**Run:**
```powershell
.\create-image.ps1
.\run-on-docker-local.ps1
.\scripts\test-dummy-data.ps1
```

Remote:
```powershell
.\run-on-docker-server.ps1 --ssh-string=<alias>
```

**Entry point:** `cmd/server/main.go`  
**API base:** `/dogan/api/v1`  
**CORS:** `DOGAN_CORS_ALLOWED_ORIGINS` defaults to local web (`:30808`) plus `https://dogan.xaigrok.ir`. Remote `--domain` appends `https://<domain>`.  
**Login:** user `armin` — plaintext in local `armin-credentials.txt` (gitignored; regenerate + update `internal/auth/credentials.go` hash if missing)
