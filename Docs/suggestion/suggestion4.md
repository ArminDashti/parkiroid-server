# Suggestion: local publish_port

Local `.armin/docker-scripts/run-on-docker-local.yaml` has no `publish_port`. Compose defaults host bind to `8080` via `DOGAN_PUBLISH_PORT`. On machines where another stack already uses 8080, local deploy fails at compose up even though the image builds.

Consider adding optional `publish_port` to the local YAML/script (same pattern as server → `DOGAN_PUBLISH_PORT`) so operators can pick a free host port without editing `docker-compose.yml`.
