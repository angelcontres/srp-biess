---
name: srp-verification
description: >-
  Use this skill to run local quality checks, compile verification, and test execution for SRP-BIESS before committing code.
---

# Fast Quality Verification Skill

Procedimiento para certificar código antes de la entrega o commit.

## 1. Verificación Rápida de Backend
Ejecuta el arnés local:
```cmd
verify.bat
```
O en Linux/Git Bash:
```bash
./verify.sh
```

## 2. Puntos de Verificación
1. **Compilación Go:** `go build ./cmd/api` debe terminar con exit code 0.
2. **Cero Violaciones de Dominio:** Ningún archivo en `internal/domain` debe contener referencias a SQL o HTTP.
3. **Servicios Docker:** `docker compose ps` debe mostrar `postgres_biess` y `redis_biess` en estado saludable (healthy/up).
