# Server inventory (agent reference)

## Irancell-T3 (`t3` / `t3-new`)

| Field | Value |
|-------|-------|
| Hostname | `t3-new` |
| SSH alias | `t3` |
| Host | `2.144.27.74` |
| User | `cloud-admin` |
| Identity | `~/.ssh/id_ed25519_cloud-admin_2-144-27-74` (also skill mentions `id_ed25519_irancell`) |
| Docker network | `t3-net` |
| Resources | ~4 GB RAM available shown; skill notes 2C/2G/30G class — verify live |

### dogan stack (2026-07-24)

| Item | Value |
|------|-------|
| Volume dir | `/home/cloud-admin/docker/dogan` |
| Containers | `dogan-api`, `dogan-pgsql`, `livekit`, `dogan-webui` (on `t3-net`) |
| Image | `dogan-api:latest`; web UI `parkiroid-web:latest` |
| HAProxy hosts | `dogan-api.xaigrok.ir` → `dogan-api:8080`; `dogan-livekit.xaigrok.ir` → `livekit:7880`; `dogan.xaigrok.ir` → `dogan-webui:80` |
| Web UI volume | `/cloud-admin/docker-volumes/parkiroid-web:/data` |

### Notes

- SSH MCP profile named `irancell` currently targets `2.144.13.55` (stale); prefer `ssh t3`
- Do not put passwords in this file when avoidable; use key auth
