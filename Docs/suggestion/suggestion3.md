# Suggestion: Prefer nginx proxy for same-origin web+API in production

CORS fixes local cross-port login (`:30808` → `:8080`), but production often works better when the web UI proxies `/dogan/api` so the browser stays same-origin (no CORS list to maintain, fewer cookie/header edge cases). Consider documenting a reverse-proxy layout for deployed stacks.
