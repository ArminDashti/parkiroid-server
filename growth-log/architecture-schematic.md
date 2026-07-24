# Architecture (schematic)

Client → dogan-api (:8080) → PostgreSQL + LiveKit on external network t3-net.
Deploy scripts build `dogan-api:latest` and `docker compose -p dogan up -d`.
