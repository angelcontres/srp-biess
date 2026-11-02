---
name: design-audit
description: Audita un archivo o carpeta del sistema de postulación BIESS contra DESIGN.md (tokens, forma, componentes, accesibilidad, contenido y prohibiciones) y reporta las violaciones en una tabla sin modificar archivos. Úsala para revisar código de interfaz antes de un commit, un PR o una entrega.
---

# Auditoría de diseño

Argumentos: `$ARGUMENTS` = ruta a auditar (archivo o carpeta). Si está vacío, o ves el texto `$ARGUMENTS` sin sustituir y el usuario no indicó una ruta, audita todo `src/` (y `index.html` si existe).

**Esta skill no modifica archivos. Solo reporta.**

## 1. Prepara

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto. Si no existe, detente y avisa.
2. Lee `tailwind.config.js` para conocer los tokens reales.
3. Lista los archivos del alcance con extensión `.tsx`, `.jsx`, `.ts`, `.js`, `.css`, `.html`. Excluye `node_modules`, `dist`, `build`, archivos generados y `tailwind.config.js`. En `src/index.css`, las definiciones de tokens dentro de `@layer base` son válidas.

## 2. Revisa

Usa búsquedas (ripgrep o equivalente) y después lee el contexto de cada coincidencia: muchas reglas dependen de dónde aparece el valor.

### 🔴 Bloqueantes

| Regla | Cómo detectarla | Esperado |
|---|---|---|
| Color hardcodeado | `#[0-9a-fA-F]{3,8}\b`, `rgb\(`, `rgba\(`, `hsl\(` en componentes | Clase de token (DESIGN.md §2) |
| Paleta por defecto de Tailwind | `\b(bg|text|border|outline|fill|stroke|from|to|via|divide|placeholder|decoration)-(slate|gray|zinc|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}` | Token equivalente |
| Tono `neutral` inexistente | `neutral-(200|300|400|500|800|950)` | `line`, `line-strong`, `line-input` o `neutral-{50,100,600,700,900}` |
| Valor arbitrario | `-\[[^\]]*(px|rem|#)[^\]]*\]` | Escala de espaciado o tipografía |
| Foco con `ring` u oculto | `focus(-visible)?:ring`, `outline-none` sin reemplazo con `outline` | `focus-visible:outline-2 …` |
| Radio fuera de escala | `rounded-(xl|2xl|3xl)`; `rounded-full` en botones, tarjetas o inputs | `rounded-md` / `rounded-lg` / `rounded` |
| Desenfoque o degradado | `backdrop-blur`, `blur-`, `bg-gradient`, `from-`, `via-` | Colores sólidos |
| Contenedor delimitado solo por sombra | Tarjeta o modal con `shadow` sin `border-line-strong`, o con `border-line` | `border border-line-strong` |
| Secondary como texto o borde | `text-secondary`, `border-secondary` | Solo la franja `h-1 bg-secondary` |
| Placeholder como etiqueta | `<input` o `<textarea` con `placeholder` y sin `label` asociado | `label` visible |
| Toggle para Sí/No | `role="switch"`, componentes `Switch`/`Toggle` en preguntas | Radio buttons en `fieldset` |
| Opción preseleccionada en declaración o consentimiento | `defaultChecked`, `checked` inicial `true` en esos grupos | Ninguna marcada |
| Estado solo con color | Badges o indicadores sin texto | Icono + texto |
| Más de un botón primario | Varios `bg-primary` de acción en la misma pantalla o modal | Uno solo |
| Spinner fuera de botón | `animate-spin` en contenedores de página | Skeleton |
| Emoji o punto medio | Rango Unicode de emojis, carácter `·` | Texto o icono de lucide |
| Mayúsculas sostenidas | `uppercase`, `tracking-wide(r|st)?` en labels o títulos | Mayúscula inicial |
| Error de backend que limpia el formulario | `reset()` o vaciado de estado en el `catch` de guardado | Mantener datos (DESIGN.md §9) |
| Guardado desde ErrorBoundary | Llamadas a la API dentro de `componentDidCatch` o del fallback | Solo borrador local |

### 🟡 Sugeridos

| Regla | Cómo detectarla | Esperado |
|---|---|---|
| Tamaño de texto menor al mínimo | `text-xs`, `text-[` | `text-sm` mínimo; `text-base` en campos |
| Mensaje de error en `text-sm` | Error con `text-sm` | `text-base` |
| Movimiento sin `motion-safe:` | `animate-`, `transition` sin prefijo | `motion-safe:` |
| Fuente monoespaciada | `font-mono` | Inter con `tabular-nums` |
| Icono sin `aria-hidden` junto a texto | Componentes de lucide sin `aria-hidden` | `aria-hidden="true"` |
| Fecha o número formateado a mano | `toLocaleString()` sin `es-EC`, plantillas de fecha manuales | `Intl` con `es-EC` y `America/Guayaquil` |
| Tuteo o términos no fijos | "tu", "olvidaste", "solicitud", "aplicación", "paso" | Usted, "postulación", "sección" |
| Botón con texto genérico | "Aceptar", "OK", "Enviar" a secas, `→` | Verbo + objeto |
| Control táctil bajo | Botones o inputs con `h-8`, `h-9`, `h-10` | `h-11` (44 px) |
| `type="number"` en cédula | `type="number"` | `inputMode="numeric"` |

## 3. Reporta

Una sola tabla, ordenada por severidad y luego por archivo:

| Sev. | Archivo | Línea | Violación | Token / patrón esperado | Valor encontrado |
|---|---|---|---|---|---|
| 🔴 | `src/pages/Login.tsx` | 42 | Color hardcodeado | `text-danger` | `text-[#B42318]` |

Después de la tabla:
- Totales: archivos revisados, 🔴 bloqueantes, 🟡 sugeridos.
- Los tres arreglos de mayor impacto.
- Coincidencias que revisaste y descartaste como falsos positivos, en una línea cada una.

No apliques cambios. Si el usuario quiere corregir, ofrece hacerlo en una petición aparte.
