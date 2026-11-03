---
name: srp-verification
description: >-
  Use this skill to run local quality checks, compile verification, and test execution for SRP-BIESS before committing code.
---

# Fast Quality Verification Skill

Procedimiento para certificar código antes de la entrega o commit.

## 1. Verificación Rápida
Ejecuta el arnés local en Windows:
```cmd
verify.bat
```
O con suite completa de pruebas:
```cmd
verify.bat --full
```
O en Linux / macOS / Git Bash:
```bash
./verify.sh
```
O con suite completa de pruebas:
```bash
./verify.sh --full
```

## 2. Puntos de Verificación Automatizados (5/5)
1. **Arquitectura Limpia (ADR-003):** Cero dependencias de infraestructura en `internal/domain/` (`database/sql`, `pgx`, `fiber`, `redis`, `net/http`).
2. **Compilación Go:** `go build ./...` y `go build ./cmd/api` deben terminar con exit code 0.
3. **Análisis Estático:** `go vet ./internal/... ./cmd/...` aprobado sin advertencias o errores.
4. **Pruebas Rápidas:** `go test ./internal/domain/... ./internal/application/...` aprobadas en memoria (< 5s). Con `--full` ejecuta `go test ./...`.
5. **Servicios Docker:** `docker compose ps` reporta PostgreSQL 16, Redis 7 y Mailpit en estado activo.
