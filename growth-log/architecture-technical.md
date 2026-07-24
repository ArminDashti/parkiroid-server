# Architecture (technical)

Go Gin API in Docker; compose wires postgres:16-alpine, livekit/livekit-server, and dogan-api on external network `t3-net`. Deploy scripts under `.armin/docker-scripts` set `API_IMAGE_TAG` and `DOCKER_NETWORK` (server also `DOGAN_PUBLISH_PORT`).
