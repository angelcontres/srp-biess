---
applyTo: "frontend/**/*.{ts,tsx,js,jsx,css,html}"
---

# Frontend React de SRP-BIESS

Antes de modificar interfaz, lee `DESIGN.md` y la skill aplicable en
`.agents/skills/`. Usa `new-page`, `new-form-section`, `new-component`,
`add-section`, `design-audit`, `responsive-check` o `token-check` según el tipo
de trabajo.

Respeta únicamente los tokens de `DESIGN.md`, WCAG 2.2 AA, los textos en usted,
los términos fijos `postulación` y `sección`, y el manejo de errores sin borrar
datos del formulario. No inventes contenido institucional; usa
`[PENDIENTE: ...]` cuando el SDD no lo defina.

Antes de concluir, ejecuta los scripts relevantes definidos en
`frontend/package.json` para lint, typecheck, test y build.
