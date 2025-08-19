# Coffee — Step 1 (Boot + Config) — Minimal

Includes only bootstrapping:
- Cobra/Viper CLI
- Gin HTTP server
- Functional options
- Context Request-ID + JSON access logs (slog)
- /v1/healthz and /v1/readyz
- Strict server timeouts + graceful shutdown
- One unit test

## Run
```bash
go mod tidy
make run
# new terminal:
curl -i http://localhost:8080/v1/healthz
curl -i http://localhost:8080/v1/readyz
```
