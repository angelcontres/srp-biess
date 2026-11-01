# GitHub Copilot Instructions — SRP-BIESS

- Follow Clean Architecture for Go Fiber (ADR-003).
- Never import database drivers or infrastructure inside `internal/domain`.
- Always generate code compatible with Go 1.22+ and React 18+.
- Keep responses compact to minimize token consumption.
