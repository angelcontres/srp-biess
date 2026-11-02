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

## 4. Diseño de interfaz

Este proyecto es el sistema de postulación al Concurso Público de Méritos y Oposición del BIESS: un servicio público digital, no un producto de consumo.

**Antes de escribir o modificar cualquier interfaz, lee `DESIGN.md` en la raíz del proyecto.** Es la fuente de verdad visual y prevalece sobre cualquier valor por defecto de Tailwind, de librerías de componentes o de costumbre. Las reglas de negocio y los textos institucionales están en el SDD (`08-diseno-interfaz-usuario.md` y los documentos que referencia); no los inventes: usa `[PENDIENTE: …]`.

Stack: React, Tailwind CSS 3.4.17 (con `theme.colors` reemplazado por los tokens de DESIGN.md) y `lucide-react` ^0.474.0.

Reglas que nunca se rompen:

- Solo clases de tokens de DESIGN.md. Ningún hex, `rgb()`, valor arbitrario (`p-[13px]`) ni paleta por defecto (`gray-*`, `slate-*`, `blue-*`…).
- Contenedores con `border border-line-strong` y sin sombra; inputs con `border-line-input`; `line` solo para divisores internos.
- Radios: botones e inputs `rounded-md`, tarjetas y modales `rounded-lg`, badges `rounded`. Nada mayor que `rounded-lg`.
- Foco con `outline` (`focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark`), nunca `ring`.
- Un solo botón primario por pantalla. Controles de 44 px de alto (`h-11`).
- Todo estado con icono + texto; nunca solo color.
- Spinner solo dentro de botones; skeleton para cargas de layout.
- Un error del servidor nunca borra los datos del formulario.
- Prohibido: emojis, el punto medio `·`, `uppercase`, `backdrop-blur`, degradados, shimmer, toggles para Sí/No, fuente monoespaciada y placeholders como etiqueta.
- Textos en usted; términos fijos "postulación" y "sección"; fechas con `Intl` `es-EC` y zona `America/Guayaquil`.
- Accesibilidad WCAG 2.2 AA como mínimo.

### Skills

Cuando la petición coincida con una de estas tareas, lee el archivo indicado
antes de empezar y sigue sus instrucciones.

| Tarea | Leer |
|---|---|
| Crear una página o ruta | `.agents/skills/new-page/SKILL.md` |
| Crear una sección del formulario | `.agents/skills/new-form-section/SKILL.md` |
| Crear un componente | `.agents/skills/new-component/SKILL.md` |
| Agregar un bloque a una página | `.agents/skills/add-section/SKILL.md` |
| Auditar contra DESIGN.md | `.agents/skills/design-audit/SKILL.md` |
| Revisar responsive | `.agents/skills/responsive-check/SKILL.md` |
| Buscar valores hardcodeados | `.agents/skills/token-check/SKILL.md` |

