# Router Module

Gin engine setup, global middleware, and `/dogan/api/v1` route registration.

## Responsibility

- Create the Gin engine with logger, recovery, and CORS
- Wire handlers and protect routes with bearer-token middleware
- Open PostgreSQL and start retention cleanup

## CORS

Browser clients (e.g. Parkiroid web on `:30808`) call the API on a different origin (`:8080`). CORS is enabled via `github.com/gin-contrib/cors`.

| Env | Default | Meaning |
|-----|---------|---------|
| `DOGAN_CORS_ALLOWED_ORIGINS` | `http://localhost:30808,http://127.0.0.1:30808,https://dogan.xaigrok.ir` | Comma-separated allowed `Origin` values |

Allowed methods: GET, POST, PUT, PATCH, DELETE, OPTIONS.  
Allowed headers: Origin, Content-Type, Authorization, Accept.

Deploy scripts merge the defaults above and, when `--domain` is set, add `https://<domain>`.

## Dependencies

- `internal/config` — listen address, secrets, CORS origins
- `internal/handlers`, `internal/middleware`, `internal/store`, `internal/auth`, `internal/livekit`
