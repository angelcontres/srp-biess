---
applyTo: "cmd/**/*.go,internal/**/*.go"
---

# Backend Go/Fiber de SRP-BIESS

Antes de trabajar, lee:

- `.agents/skills/srp-clean-backend/SKILL.md`
- `.agents/skills/srp-verification/SKILL.md`

Aplica ADR-003: dominio puro, casos de uso en `internal/application`,
adaptadores en `internal/infrastructure`, HTTP en `internal/interfaces/http` y
composición manual en `cmd/api/main.go`. No introduzcas drivers de base de datos
ni Fiber en `internal/domain`.

Después de cambiar código, ejecuta como mínimo:

```bash
go build ./cmd/api
go test -v ./...
```

Si el cambio afecta la integración local, ejecuta también `./verify.sh`.
