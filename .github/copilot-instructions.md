# GitHub Copilot Instructions — SRP-BIESS

## Alcance

Estas instrucciones son específicas para GitHub Copilot (CLI, coding agent y
agentes personalizados). No sustituyen ni modifican las configuraciones de
otros agentes.

## Resolución de skills

Antes de leer, escribir, revisar o probar código, identifica la skill aplicable
y lee el `SKILL.md` completo:

| Tipo de trabajo | Skill |
|---|---|
| Página o ruta | `.agents/skills/new-page/SKILL.md` |
| Sección del formulario | `.agents/skills/new-form-section/SKILL.md` |
| Componente UI | `.agents/skills/new-component/SKILL.md` |
| Bloque en una página | `.agents/skills/add-section/SKILL.md` |
| Auditoría visual | `.agents/skills/design-audit/SKILL.md` |
| Responsive | `.agents/skills/responsive-check/SKILL.md` |
| Tokens hardcodeados | `.agents/skills/token-check/SKILL.md` |
| Backend Go/Fiber | `.agents/skills/srp-clean-backend/SKILL.md` |
| Verificación | `.agents/skills/srp-verification/SKILL.md` |
| Optimización de salida | `.agents/skills/srp-token-saver/SKILL.md` |

Si una tarea combina áreas, carga todas las skills aplicables. No inventes una
skill ni sustituyas una instrucción específica por un resumen.

## Reglas de implementación

- Sigue Clean Architecture (ADR-003) para Go Fiber.
- `internal/domain` solo contiene Go puro: nunca importa `database/sql`, `pgx`,
  Fiber ni infraestructura.
- Usa inyección manual de dependencias en `cmd/api/main.go`.
- Genera código compatible con Go 1.22+ y React 18+.
- En interfaces, lee `DESIGN.md` antes de modificar código y respeta sus tokens,
  accesibilidad y textos institucionales.
- Haz cambios quirúrgicos; no reviertas cambios preexistentes ni edites archivos
  fuera del alcance.
- Los errores del servidor no deben borrar datos de formularios.
- No inventes reglas de negocio: usa el SDD o marca `[PENDIENTE: ...]`.

## Verificación obligatoria

Ejecuta los checks relevantes antes de concluir:

- Backend: `go build ./cmd/api` y `go test -v ./...`.
- Frontend: scripts existentes de `lint`, `typecheck`, `test` y `build` en
  `frontend/package.json`.
- Verificación integrada: `./verify.sh` en Linux o `verify.bat` en Windows.

Reporta solo los errores relevantes, con archivo, línea y mensaje principal.
Mantén las respuestas compactas y menciona el comando exacto ejecutado.
