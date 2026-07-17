# Module: Docker deploy scripts

## Responsibility

Build and deploy the dogan-server Compose stack (API + PostgreSQL + LiveKit) locally or over SSH.

## Entry points

| Script | Role |
|--------|------|
| `create-image.ps1` | Build (and optionally export) `dogan-server` image |
| `run-on-docker-local.ps1` | Local daemon deploy |
| `run-on-docker-server.ps1` | Remote deploy; requires `--ssh-string=<alias>` |
| `run-on-docker.ps1` | Shared engine invoked by local/server scripts |

Wrappers call the engine with `-RemainingArguments @(...)` so `--flag=value` strings are not mis-bound to typed parameters (e.g. `SshString`). Use `--mode=local|server` or `-DeployMode` on the engine; do not alias a parameter as `Mode`/`mode` (PowerShell case-insensitive name clash).

## Defaults (when flags are null)

| Flag | Resolved value |
|------|----------------|
| `--delete-image` / `--delete-volume` | `no` |
| `--internal-port` (local/server) | `8080` from `.docker/stack.manifest.json` |
| Postgres host port (local) | Random free port `30000–32767` (avoids clash with other stacks on 5432) |
| `--volume-dir` | `<USER>/docker/dogan-server` |
| `--volume-name` | `dogan-server-volume` |
| `--network-name` | `dogan-net` from manifest |

## CORS on deploy

Compose sets `DOGAN_CORS_ALLOWED_ORIGINS` (local web + `https://dogan.xaigrok.ir`).  
`run-on-docker.ps1` / `--domain=<host>` also passes `https://<host>` into that env var on compose up.

## Dependencies

- `Dockerfile`, `docker-compose.yml`, `livekit.yaml`
- `.docker/stack.manifest.json`
- Docker CLI; remote also needs SSH config alias + `scp`
