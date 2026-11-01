---
name: srp-token-saver
description: >-
  Use this skill to optimize prompt construction, error reporting, and code diffs to minimize token usage across all LLMs in SRP-BIESS.
---

# Token Optimization Skill

Pautas para ahorrar tokens al trabajar con cualquier IA en este proyecto:

## 1. Reglas de Salida Concisa
* No generar código repetitivo ni comentarios obvios.
* Usar diffs o bloques de modificación enfocados en lugar de reescribir archivos enteros.
* Omitir cortesías conversacionales.

## 2. Compresión de Errores
Cuando una compilación falle:
* No enviar el log completo de 200 líneas.
* Enviar únicamente las líneas que comiencen con `error:`, `FAIL:`, o `undefined:`.
* Ejemplo:
  ```text
  cmd/api/main.go:25:12: undefined: handler.NewAuthHandler
  ```
Esto ahorra más de 2,000 tokens por consulta.
