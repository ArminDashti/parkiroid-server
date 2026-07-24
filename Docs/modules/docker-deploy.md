# Module: Docker deploy scripts

## Responsibility

Build and deploy the dogan Compose stack (API + PostgreSQL + LiveKit) locally or over SSH via YAML-only scripts under `.armin/docker-scripts/`.

## Entry points

| Script | Role |
|--------|------|
| `.armin/docker-scripts/run-on-docker-local.ps1` | Local daemon deploy (reads sibling YAML) |
| `.armin/docker-scripts/run-on-docker-server.ps1` | Remote SSH deploy (reads sibling YAML) |
| `create-image.ps1` | Legacy image build helper at repo root |
| `run-on-docker*.ps1` (repo root) | Older CLI-flag wrappers; prefer `.armin/docker-scripts/` |

Behavior is controlled only by YAML — no CLI `--` flags on the `.armin` scripts.

## YAML settings

| Key | Local | Server | Notes |
|-----|-------|--------|-------|
| `stack_name` | `dogan` | `dogan` | Compose project name (`-p`) |
| `image_tag` | `dogan-api:latest` | same | Passed as `API_IMAGE_TAG` |
| `compose_file` | `../../docker-compose.yml` | same | Relative to `.armin/docker-scripts/` |
| `dockerfile` | `../../Dockerfile` | same | Relative to `.armin/docker-scripts/` |
| `docker_network` | `t3-net` | same | External network; created if missing |
| `internal_port` | empty | empty | Unused by this compose (container listens on 8080) |
| `publish_port` | `30251` | empty | Maps to `DOGAN_PUBLISH_PORT`; local avoids clash with other stacks on 8080; empty on server clears host bind |
| `ssh` / `volume_dir` | — | `ssh t3` / `/home/cloud-admin/docker/dogan` | Irancell-T3 (`t3-new`) |
| `build_image_on` | — | `local` | Prefer local build+upload; remote GOPROXY returns 403 on this host |
| `delete_image` | — | `yes` | Teardown runs before image load so the new image is not deleted |

## Compose env overrides

Scripts set env vars that match `docker-compose.yml`:

- `API_IMAGE_TAG`
- `DOCKER_NETWORK`
- `DOGAN_PUBLISH_PORT` (local when `publish_port` set; always on server)

## CORS on deploy

Compose sets `DOGAN_CORS_ALLOWED_ORIGINS` (local web + `https://dogan.xaigrok.ir`).

## Dependencies

- `Dockerfile`, `docker-compose.yml`, `livekit.yaml`
- Docker CLI; remote also needs SSH alias (or `sshpass` for password mode) + `scp`
- Local publish defaults to host port `30251` via YAML (`lexmora-api` already owns `8080` on shared hosts)
