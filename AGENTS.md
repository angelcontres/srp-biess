# REGLAS MAESTRAS DE DESARROLLO CON IA — PROYECTO SRP-BIESS
Este archivo es leído automáticamente por agentes de IA (Antigravity, Cursor, Claude, Copilot, ChatGPT).

## 1. Contexto del Proyecto
* **Nombre:** Sistema de Registro y Postulación al Directorio del BIESS (SRP-BIESS).
* **Backend:** Go 1.22+ / 1.23+ con Fiber v2.
* **Frontend:** React 18+ (SPA), Vite, Tailwind CSS.
* **Persistencia & Caché:** PostgreSQL 16 (`srp_biess`), Redis 7 (`db 0`), Mailpit (`smtp:1025`, `web:8025`).
* **Catálogo de Tickets:** [`Docs/TICKETS.md`](./docs/TICKETS.md) y Especificaciones en [`Docs/SDD/`](./docs/SDD/).

---

## 2. Reglas Estrictas de Arquitectura (ADR-003)
1. **Separación de Capas en Go:**
   * `internal/domain/`: Entidades y puertos (interfaces puras). **NUNCA** debe importar `database/sql`, `pgx`, `fiber` ni infraestructura.
   * `internal/application/`: Casos de uso y orquestación de lógica de negocio.
   * `internal/infrastructure/`: Repositorios SQL (`pgxpool`), caché Redis, clientes externos.
   * `internal/interfaces/http/`: Handlers Fiber, DTOs y middlewares.
   * `cmd/api/main.go`: Composition Root con inyección de dependencias **manual** (sin frameworks de reflexión).
2. **Validación Módulo 10:**
   * La validación de cédulas ecuatorianas y RUC debe cumplir con el algoritmo Módulo 10 según `Docs/SDD/14-decisiones-de-arquitectura.md`.

---

## 3. Protocolo de Ahorro Máximo de Tokens
* **Concisión Extrema:** No respondas con saludos ni introducciones largas. Ve directo al código o respuesta técnica.
* **Ediciones Quirúrgicas:** Muestra solo las funciones o archivos que cambiaron. No reescribas archivos gigantes innecesariamente.
* **Verificación Local:** Antes de dar una tarea por terminada, indica el comando exacto para verificarla (`go build ./cmd/api` o `verify.bat`).
