# Directory Tree

```
.
├── build-docker-image.ps1       # PowerShell Docker build script
├── cmd/
│   ├── issue-token/main.go      # Generate embedded API token (dg_*)
│   └── server/main.go           # HTTP server entry point
├── docker-compose.yml           # postgres + livekit + dogan-server
├── Dockerfile                   # Multi-stage Go build image
├── docs/                        # Project documentation (this folder)
├── endpoints.md                 # Short public endpoint table
├── internal/
│   ├── auth/                    # JWT, bcrypt login, embedded token
│   ├── config/config.go         # DOGAN_* environment configuration
│   ├── handlers/                # REST HTTP handlers
│   ├── livekit/                 # LiveKit token and room naming
│   ├── middleware/auth.go       # Bearer token middleware
│   ├── models/models.go         # Request/response DTOs
│   ├── router/router.go         # Route registration
│   └── store/                   # PostgreSQL persistence layer
├── livekit.yaml                 # LiveKit server config
├── README.md                    # Quick start guide
└── scripts/
    └── test-dummy-data.ps1      # Docker integration smoke test
```
