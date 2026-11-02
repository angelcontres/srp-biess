---
name: new-component
description: Crea un componente de interfaz reutilizable para el sistema de postulación BIESS usando solo los tokens, radios, bordes, estados y reglas de accesibilidad de DESIGN.md. Úsala cuando se pida un botón, campo, tarjeta, alerta, badge, modal u otro componente nuevo.
---

# Crear un componente

Argumentos: `$ARGUMENTS` = nombre del componente (por ejemplo `FileUploadField`). Si ves el texto `$ARGUMENTS` sin sustituir, toma el nombre del mensaje del usuario; si no hay ninguno, pregúntalo.

## 1. Lee las fuentes

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto. Si no existe, detente y avisa.
2. Busca el componente en DESIGN.md §7. Si está especificado allí, esa especificación **es obligatoria**: estructura, clases, estados, textos e iconos.
3. Si el componente **no** aparece en DESIGN.md, no lo inventes. Comprueba si se puede componer con los existentes. Si de verdad hace falta uno nuevo, describe al usuario qué propones (propósito, variantes, estados) y espera su confirmación antes de escribirlo.
4. Revisa los componentes existentes (`src/components`, `src/ui`…) y copia sus convenciones:
   - TypeScript o JavaScript.
   - Si usan `forwardRef`.
   - La utilidad para combinar clases (`cn`, `clsx`, `twMerge`).
   - Cómo se definen variantes (por ejemplo `cva`) y cómo se exportan (archivo índice).
   - Si hay tests o stories junto a cada componente.
5. Verifica que el componente no exista ya con otro nombre.

## 2. Reglas de construcción

**Tokens (DESIGN.md §2, §3, §13)**
- Solo clases definidas en `tailwind.config.js`: `primary`, `primary-dark`, `secondary`, `background`, `surface`, `muted`, `neutral-{50,100,600,700,900}`, `line`, `line-strong`, `line-input`, `info`, `success`, `warning`, `danger` y sus `-50`.
- Ningún hex, `rgb()`, valor arbitrario ni paleta por defecto de Tailwind.
- Espaciado de la escala 4, 8, 12, 16, 24, 32, 48, 64 px.
- Tipografía de la escala de DESIGN.md: mínimo `text-sm` y nunca para errores; campos e instrucciones con `text-base`.

**Forma (DESIGN.md §5)**
- Botones e inputs `rounded-md`; tarjetas y modales `rounded-lg`; badges y checkbox `rounded`; `rounded-full` solo para radio, spinner y círculos del stepper.
- Contenedores con `border border-line-strong`, sin sombra. Sombra solo para tooltip (`shadow-sm`), menú (`shadow-md`) y modal (`shadow-lg`).
- Controles interactivos con 44 px de alto como mínimo (`h-11`).

**Estados obligatorios**

| Estado | Implementación |
|---|---|
| Default | Límite visible ≥ 3:1. |
| Hover | Cambio de color instantáneo, sin mover ni escalar. |
| Focus | `focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark` (`outline-white` sobre fondo `primary`). **Nunca `ring`.** |
| Active | Cambio de tono. |
| Disabled | `aria-disabled` o `disabled`, `bg-neutral-100 text-muted border-line-input`, y una prop o slot para explicar el motivo. |

Según el tipo de componente, añade:
- Si ejecuta una acción, estado **loading**: `LoaderCircle` con `motion-safe:animate-spin` + texto, ancho estable y sin segundo clic.
- Si es un campo, estados **error** (borde de 2 px `danger` + mensaje con `CircleAlert` vinculado por `aria-describedby`, `aria-invalid`), **read-only** y **verificado** (DESIGN.md §7.4).
- Si muestra estado, **icono + texto**, nunca solo color.

**Accesibilidad (DESIGN.md §10)**
- Elemento HTML nativo siempre que exista (`button`, `input`, `select`, `dialog`…).
- Operable con teclado; nombre, rol y valor correctos.
- `label` visible asociado (nunca `placeholder` como etiqueta); `fieldset` + `legend` para grupos.
- Iconos de `lucide-react` con `aria-hidden="true"` cuando acompañan texto; los botones de solo icono llevan `aria-label`.
- Mensajes dinámicos en `aria-live="polite"` o `role="alert"` según DESIGN.md.
- Funciona en `forced-colors`: bordes reales, foco con `outline`, iconos en `currentColor`.

**Movimiento (DESIGN.md §8)**
- Solo con `motion-safe:`. Sin animaciones de entrada, shimmer, rebotes ni escalado.

**Prohibido (DESIGN.md §12)**
- Emojis, `·`, `uppercase`, `tracking-*` en etiquetas, `backdrop-blur`, degradados, `rounded-xl` o mayores, botones píldora, toggles para Sí/No, fuente monoespaciada y flechas `→` en textos de botón.

## 3. API y exportación

- Props tipadas y exportadas (`export type <Nombre>Props`).
- Acepta `className` y lo combina con la utilidad del proyecto, sin permitir que sustituya los estados de foco.
- Reenvía `ref` si los demás componentes lo hacen.
- Exporta el componente y sus tipos desde el índice de componentes, si existe.
- Si el proyecto tiene tests o stories por componente, crea los equivalentes cubriendo todos los estados.

## 4. Verifica e informa

1. Ejecuta `lint`, `typecheck` y `test` si existen, y corrige los fallos.
2. Busca en el archivo nuevo: `#[0-9a-fA-F]{3,8}`, `\[[0-9.]+px\]`, `ring-`, `rounded-(xl|2xl|3xl)`, `shadow` en contenedores, `backdrop-blur`, `uppercase`. No debe quedar ninguno.
3. Informa: archivo creado, props, estados cubiertos, dónde se exporta y si algo quedó pendiente de confirmar.
