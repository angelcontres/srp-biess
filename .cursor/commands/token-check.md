<!-- Busca en el código del sistema de postulación BIESS colores, medidas, radios, sombras y fuentes hardcodeados o fuera de la configuración, y propone el token de DESIGN.md que debe reemplazarlos, sin aplicar cambios sin confirmación. Úsala antes de un commit o cuando se sospeche que se usan valores sueltos. -->

# Verificación de tokens

Sin argumentos. Alcance: `src/` e `index.html`. Excluye `node_modules`, `dist`, `build`, archivos generados y `tailwind.config.js`.

**No apliques cambios sin confirmación del usuario.**

## 1. Prepara

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto, en especial el front matter YAML, §2, §3, §5 y §13. Si no existe, detente y avisa.
2. Lee `tailwind.config.js` y compáralo con DESIGN.md §13. Si difieren, repórtalo primero: la configuración es la fuente de verdad de las clases disponibles.

## 2. Busca

| Categoría | Patrón | Contexto a revisar |
|---|---|---|
| Hex | `#[0-9a-fA-F]{3,8}\b` | Clases arbitrarias, `style`, CSS, SVG inline (`fill`, `stroke`). |
| Funciones de color | `rgba?\(`, `hsla?\(` | Igual. |
| Paletas por defecto | `-(slate|gray|zinc|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}` | No existen en la configuración: no generan estilo. |
| `neutral` inexistente | `neutral-(200|300|400|500|800|950)` | No existen en la configuración. |
| Medidas arbitrarias | `-\[[0-9.]+(px|rem|em)\]` | Padding, margin, gap, width, height, text, rounded. |
| px en estilos | `[0-9]+px` en `style={{…}}` y CSS | Excepto `scroll-padding-top` y definiciones de tokens. |
| Fuentes | `font-family`, `fontFamily`, `font-\[`, `font-mono`, `font-serif` | Solo Inter vía `font-sans`. |
| Sombras | `shadow-\[`, `box-shadow`, `shadow-(xl|2xl|inner)` | Solo `shadow-sm`, `shadow-md`, `shadow-lg` y en sus usos. |
| Radios | `rounded-\[`, `rounded-(xl|2xl|3xl)`, `border-radius` | Escala de §5. |
| Pesos | `font-(thin|extralight|light|extrabold|black)` | Solo 400, 500, 600, 700. |

## 3. Tabla de reemplazos

**Colores**

| Valor encontrado | Token |
|---|---|
| `#0B5CAD` | `primary` / `info` |
| `#083F78` | `primary-dark` |
| `#F2B705` | `secondary` (solo franja bajo el header) |
| `#F8FAFC` | `background` / `neutral-50` |
| `#FFFFFF`, `#fff`, `white` en superficies | `surface` |
| `#0F172A` | `neutral-900` |
| `#334155` | `neutral-700` |
| `#475569` | `muted` / `neutral-600` |
| `#F1F5F9` | `neutral-100` |
| `#E2E8F0` | `line` (solo divisores internos) |
| `#7F8DA1` | `line-strong` |
| `#64748B` | `line-input` |
| `#1E7B45` / `#ECFDF3` | `success` / `success-50` |
| `#B45309` / `#FFFBEB` | `warning` / `warning-50` |
| `#B42318` / `#FEF3F2` | `danger` / `danger-50` |
| `#EFF6FF` | `info-50` |
| Cualquier otro color | Sin token: elegir el token por **función** (texto, borde, estado) y señalarlo como decisión a confirmar. |

Para las paletas por defecto, propón el token por función: `gray-500` en un borde de input pasa a `border-line-input`; `gray-500` en texto pasa a `text-muted`; `blue-600` en un botón pasa a `bg-primary`.

**Medidas**

| Encontrado | Token |
|---|---|
| 4, 8, 12, 16, 24, 32, 48, 64 px | `1`, `2`, `3`, `4`, `6`, `8`, `12`, `16` de la escala de Tailwind |
| Otro valor de espaciado | El más cercano de la escala, indicándolo |
| 44 px de alto | `h-11` |
| Radio 4 / 6 / 8 px | `rounded` / `rounded-md` / `rounded-lg` |
| Radio mayor | `rounded-lg` (tarjetas) o `rounded-md` (controles) |
| Texto 14 / 16 / 18 / 20 / 24 / 30 px | `text-sm` / `text-base` / `text-lg` / `text-xl` / `text-2xl` / `text-3xl` |
| Texto menor de 14 px | `text-sm` |
| Ancho máximo 1024 / 1280 px | `max-w-5xl` / `max-w-7xl` |

## 4. Reporta

| Archivo | Línea | Valor encontrado | Categoría | Token sugerido | Fix propuesto |
|---|---|---|---|---|---|
| `src/components/Card.tsx` | 12 | `shadow-[0_2px_8px_rgba(0,0,0,.1)]` | Sombra | Ninguna (contenedor) | `border border-line-strong`, quitar sombra |

Al final:
- Totales por categoría.
- Valores sin token equivalente que requieren decisión del usuario.
- La pregunta: "¿Aplico los reemplazos propuestos?". Solo con un sí explícito, aplica los cambios y vuelve a ejecutar las búsquedas para confirmar que no queda ninguno.
