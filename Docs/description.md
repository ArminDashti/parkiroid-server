# Dogan Server

Dogan is a Go REST API for a personal Android + web monitoring stack. It stores camera frames, Android telemetry, app settings, NCNN AI model metadata and binaries, and phone actions in PostgreSQL, and issues LiveKit tokens for WebRTC streaming between Android and the web client.

**Stack:** Go 1.26, Gin, PostgreSQL 16, LiveKit, Docker Compose.

**Run (preferred):**
```powershell
.\.armin\docker-scripts\run-on-docker-local.ps1
.\scripts\test-dummy-data.ps1
```

Local API URL after deploy: `http://localhost:30251/dogan/api/v1/health` (`publish_port` in `run-on-docker-local.yaml`; avoids host `8080` clashes).

Remote (Irancell-T3 — YAML already set for `ssh t3`):
```powershell
.\.armin\docker-scripts\run-on-docker-server.ps1
```

Production HAProxy hosts on t3: `https://dogan-api.xaigrok.ir`, `wss://dogan-livekit.xaigrok.ir`, web `https://dogan.xaigrok.ir` (separate `dogan-webui` container).

Legacy root wrappers (`.\run-on-docker-local.ps1`, `.\create-image.ps1`) still exist; prefer `.armin/docker-scripts/`.

**Entry point:** `cmd/server/main.go`  
**API base:** `/dogan/api/v1`  
**CORS:** `DOGAN_CORS_ALLOWED_ORIGINS` defaults to local web (`:30808`) plus `https://dogan.xaigrok.ir`. Remote `--domain` appends `https://<domain>`.  
**Login:** user `armin` — plaintext in local `armin-credentials.txt` (gitignored; regenerate + update `internal/auth/credentials.go` hash if missing)
