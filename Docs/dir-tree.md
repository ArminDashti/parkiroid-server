# Directory Tree

```
.
├── create-image.ps1             # Build dogan-server Docker image
├── build-docker-image.ps1       # Compat wrapper → create-image.ps1
├── run-on-docker-local.ps1      # Local Docker Compose deploy
├── run-on-docker-server.ps1     # Remote Docker deploy over SSH
├── run-on-docker.ps1            # Deploy engine used by local/server scripts
├── cmd/
│   ├── issue-token/main.go      # Generate embedded API token (dg_*)
│   └── server/main.go           # HTTP server entry point
├── docker-compose.yml           # postgres + livekit + dogan-server
├── Dockerfile                   # Multi-stage Go build image
├── .docker/
│   └── stack.manifest.json      # Stack/image/network defaults
├── docker/
│   └── agent-export/            # Optional agent-export sidecar image
├── Docs/                        # Project documentation (this folder)
├── endpoints.md                 # Short public endpoint table
├── internal/
│   ├── auth/                    # JWT, bcrypt login, embedded token
│   ├── config/config.go         # DOGAN_* env config (incl. CORS origins)
│   ├── handlers/                # REST HTTP handlers
│   ├── livekit/                 # LiveKit token and room naming
│   ├── middleware/auth.go       # Bearer token middleware
│   ├── models/models.go         # Request/response DTOs
│   ├── router/router.go         # Routes + CORS middleware
│   └── store/                   # PostgreSQL persistence layer
│       ├── model_files.go       # NCNN paths, SHA-256, file checks
├── livekit.yaml                 # LiveKit server config
├── README.md                    # Quick start guide
└── scripts/
    ├── register-models.ps1      # Scan DOGAN_MODELS_DIR and POST /ai-models
    └── test-dummy-data.ps1      # Docker integration smoke test
```
