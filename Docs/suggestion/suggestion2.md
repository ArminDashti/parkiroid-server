# Suggestion: Prefer the three Docker entry scripts

Use `create-image.ps1`, `run-on-docker-local.ps1`, and `run-on-docker-server.ps1` as the primary Docker workflow. `build-docker-image.ps1` is a thin wrapper; `run-on-docker.ps1` remains the shared deploy engine. Consider removing the wrappers after callers migrate.
