# Claude Code Guidelines — SRP-BIESS

## Build and Test Commands
- Run backend build: `go build ./cmd/api`
- Run Go tests: `go test -v ./...`
- Fast verification harness: `verify.bat` or `./verify.sh`
- Docker services: `docker compose up -d`

## Architecture Rules (ADR-003)
- Strict layer separation in Go: `cmd/api`, `internal/domain`, `internal/application`, `internal/infrastructure`, `internal/interfaces/http`.
- The domain layer (`internal/domain`) must be 100% pure Go: NO imports of `database/sql`, `pgx`, or HTTP frameworks.
- Manual dependency injection in `cmd/api/main.go`.

## Token Optimization
- Be extremely concise. Avoid repetitive explanations.
- Prefer targeted patches over full file rewrites.
