# Tokens de Diseño

Este archivo define los tokens visuales, sus roles semánticos y la configuración base de estilos. Los valores brutos en YAML residen exclusivamente en el front matter de `DESIGN.md`.

---

## 1. Color

### Roles y pares permitidos

| Uso | Clase Tailwind | Contraste |
|---|---|---|
| Texto principal | `text-neutral-900` | 17:1 |
| Texto secundario, ayudas, metadatos | `text-muted` | 7.58:1 (AAA) |
| Instrucciones | `text-neutral-700` | > 10:1 |
| Enlaces | `text-primary`, hover `text-primary-dark` | 6.67:1 |
| Botón primario | `bg-primary text-white`, hover `bg-primary-dark` | 6.67:1 |
| Error | `text-danger` sobre `bg-surface` o `bg-danger-50` | ≥ 6:1 |
| Advertencia | `text-warning` sobre `bg-warning-50` | 4.84:1 |
| Éxito | `text-success` sobre `bg-success-50` | 5.01:1 |
| Aviso institucional | `text-neutral-900` sobre `bg-info-50`, borde `border-primary` | — |
| Franja institucional | `bg-secondary`, 4 px, bajo el header | decorativa |

### Bordes (regla crítica)

| Clase | Valor | Cuándo |
|---|---|---|
| `border-line-strong` | `#7F8DA1` | **Todo contenedor**: tarjeta de login y registro, tarjetas de ítem, tarjetas de cargo, modales, barra de sección, menús. |
| `border-line-input` | `#64748B` | Bordes de inputs, selects, textareas, checkboxes, radios. |
| `border-line` | `#E2E8F0` | Solo divisores **dentro** de un contenedor (filas de lista, separación de bloques internos). |

Motivo: el contraste entre `background` y `surface` es 1.1:1. Sin un borde de 3:1 o más, la tarjeta central del login desaparece. La sombra no resuelve esto y no se usa para ello.

### Prohibido

- `secondary` (`#F2B705`) como texto, borde, icono o indicador de estado: da 1.82:1 sobre blanco.
- Usar las paletas por defecto de Tailwind (`slate-*`, `gray-*`, `blue-*`…). El `tailwind.config.js` las elimina (ver sección de Tailwind más abajo).
- Degradados, transparencias decorativas y colores fuera de tokens.
- Comunicar un estado solo con color: siempre icono y texto.
- Usar `danger` para resultados administrativos (no elegible, nacionalidad): estos van con estilo institucional `info`.

---

## 2. Tipografía

- Una sola familia: **Inter**, autoalojada (WOFF2, subconjunto latino, `font-display: swap`). Debe ser legible con la fuente de sistema si Inter no carga.
- Pesos permitidos: 400, 500, 600, 700. Nunca 800 ni 300.

| Rol | Clases | Uso |
|---|---|---|
| Display | `text-3xl font-bold leading-tight` | Solo el título de la landing. |
| H1 | `text-xl sm:text-2xl font-bold leading-tight` | Título de sección (va dentro de la barra sticky) y de pantallas. |
| H2 | `text-lg font-semibold` | Bloques dentro de una sección. |
| H3 | `text-base font-semibold` | Subgrupos, título de tarjeta. |
| Cuerpo | `text-base leading-normal` | Párrafos, textos legales, valores. |
| Label | `text-base font-medium` | Etiquetas de campo. |
| Caption | `text-sm text-muted` | Ayudas, metadatos ("Sección 3 de 10", "Último guardado"). |

Reglas:

- **Campos, textos legales e instrucciones: 16 px mínimo.** `text-sm` (14 px) es el mínimo absoluto y no se usa para errores ni instrucciones.
- Texto corrido con `max-w-prose` (≤ 75 caracteres).
- Cédulas, códigos, fechas y horas con `tabular-nums`. Nunca con fuente monoespaciada.
- Mayúscula inicial en todo. Prohibidos `uppercase` y `tracking-wide` en labels y títulos.
- No destacar una palabra suelta de un título con otro color o en cursiva.

---

## 3. Espaciado

| Relación | Valor | Clase |
|---|---|---|
| Label → ayuda | 4 px | `mt-1` |
| Ayuda → input | 8 px | `mt-2` |
| Input → error | 8 px | `mt-2` |
| Campo → campo | 24 px | `space-y-6` |
| Bloque → bloque | 32–48 px | `space-y-8` / `space-y-12` |
| Padding de tarjeta | 16 / 24 px | `p-4 sm:p-6` |
| Gutter lateral móvil | 16 px | `px-4` |

Solo valores de la escala: 4, 8, 12, 16, 24, 32, 48, 64 px. Nada de `p-[13px]`.

---

## 4. Forma: radios, bordes y elevación

| Elemento | Radio | Borde | Sombra |
|---|---|---|---|
| Botón | `rounded-md` | según variante | ninguna |
| Input, select, textarea | `rounded-md` | `border border-line-input` | ninguna |
| Checkbox | `rounded` (4 px) | `border-line-input` | ninguna |
| Radio | `rounded-full` | `border-line-input` | ninguna |
| Tarjeta (login, ítem, cargo) | `rounded-lg` | `border border-line-strong` | **ninguna** |
| Barra de sección sticky | ninguno | `border-b border-line-strong` | **ninguna** |
| Badge | `rounded` (4 px) | ninguno | ninguna |
| Alerta | `rounded-md` | `border border-{tono}` + `border-l-4` | ninguna |
| Tooltip | `rounded-md` | `border-line-strong` | `shadow-sm` |
| Menú desplegable | `rounded-md` | `border-line-strong` | `shadow-md` |
| Modal (≥ 640 px) | `rounded-lg` | `border-line-strong` | `shadow-lg` |
| Modal (< 640 px) | `rounded-none` | ninguno | ninguna (pantalla completa) |
| Overlay de modal | — | — | `bg-neutral-900/60`, **sin `backdrop-blur`** |

Prohibidos `rounded-xl`, `rounded-2xl`, `rounded-3xl` y botones con `rounded-full`.

---

## 5. Iconografía (`lucide-react` ^0.474.0)

- Tamaño por defecto 20 px (`size={20}`), 16 px en badges y captions, `strokeWidth={2}`.
- Color heredado (`currentColor`), para que funcione en `forced-colors`.
- Los iconos que acompañan un texto llevan `aria-hidden="true"`. Un botón que solo tiene icono lleva `aria-label`, y solo se permite para cerrar (`X`).
- Nunca un icono solo para comunicar estado: siempre con texto.

| Significado | Icono |
|---|---|
| Completado, éxito, guardado | `CircleCheck` |
| Error | `CircleAlert` |
| Advertencia | `TriangleAlert` |
| Aviso institucional, información | `Info` |
| Dato verificado | `ShieldCheck` |
| No disponible, bloqueado | `Lock` |
| Cargando (solo en botones) | `LoaderCircle` |
| Documento | `FileText` |
| Subir archivo | `Upload` |
| Ver | `Eye` |
| Editar | `Pencil` |
| Eliminar | `Trash2` |
| Agregar | `Plus` |
| Descargar | `Download` |
| Copiar código | `Copy` |
| Cerrar | `X` |
| Desplegar | `ChevronDown` |
| Cerrar sesión | `LogOut` |
| Reintentar | `RotateCw` |
| Mostrar/ocultar contraseña | `Eye` / `EyeOff` |

No se usan iconos decorativos en títulos, tarjetas ni listas.

---

## 6. Configuración de Tailwind 3.4.17

`theme.colors` se **reemplaza** (no `extend`), para que las paletas por defecto no existan y el agente no pueda usarlas.

```js
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      white: '#FFFFFF',
      primary: { DEFAULT: '#0B5CAD', dark: '#083F78' },
      secondary: '#F2B705',
      background: '#F8FAFC',
      surface: '#FFFFFF',
      muted: '#475569',
      neutral: {
        50: '#F8FAFC',
        100: '#F1F5F9',
        600: '#475569',
        700: '#334155',
        900: '#0F172A',
      },
      line: { DEFAULT: '#E2E8F0', strong: '#7F8DA1', input: '#64748B' },
      info: { DEFAULT: '#0B5CAD', 50: '#EFF6FF' },
      success: { DEFAULT: '#1E7B45', 50: '#ECFDF3' },
      warning: { DEFAULT: '#B45309', 50: '#FFFBEB' },
      danger: { DEFAULT: '#B42318', 50: '#FEF3F2' },
    },
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', '"Segoe UI"', 'Roboto', 'sans-serif'],
      },
      boxShadow: {
        sm: '0 1px 2px rgba(15,23,42,0.08)',
        md: '0 4px 8px rgba(15,23,42,0.12)',
        lg: '0 12px 24px rgba(15,23,42,0.18)',
      },
      // Radios: se usan los valores por defecto de Tailwind 3.4
      // rounded = 4px (badge, checkbox), rounded-md = 6px (botón, input), rounded-lg = 8px (tarjeta, modal)
    },
  },
  plugins: [],
};
```

```css
/* src/index.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  html {
    scroll-padding-top: 6rem; /* ajustar a la altura real de la barra de sección */
  }
  body {
    @apply bg-background text-neutral-900 font-sans text-base leading-normal antialiased;
  }
  a {
    @apply text-primary underline underline-offset-4 hover:text-primary-dark;
  }
  :focus-visible {
    @apply outline outline-2 outline-offset-2 outline-primary-dark;
  }
}
```
