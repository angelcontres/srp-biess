# Layout y Estructura Responsive

Este archivo define la cuadrícula, los breakpoints, los wireframes para móvil y escritorio, y las reglas de posicionamiento sticky y scroll.

---

## 1. Breakpoints (Tailwind 3.4 por defecto)

| Rango | Prefijo | Comportamiento |
|---|---|---|
| < 640 px | — | Una columna. Botones del final a ancho completo, apilados, primario arriba. Modales a pantalla completa. Tablas como tarjetas. |
| 640–1023 px | `sm:` `md:` | Dos columnas solo para campos cortos relacionados (provincia/cantón). Stepper horizontal bajo el header. |
| ≥ 1024 px | `lg:` | Stepper como barra lateral izquierda; contenido a la derecha, `max-w-5xl`. |

---

## 2. Estructura de pantalla de formulario

### Móvil (< 640 px)

```text
┌────────────────────────────────────┐
│ Header (bg-primary, no sticky)     │
├────────────────────────────────────┤ ← franja secondary 4 px
│ Sección 3 de 10  [Guardar progreso]│ ← BARRA DE SECCIÓN (sticky)
│ Familiares                         │
│▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░│ ← línea de progreso 4 px
├────────────────────────────────────┤
│ Instrucciones                      │
│ Campos…                            │
│                                    │
│ Último guardado: hoy, 14:32        │
│ [      Guardar y continuar       ] │
│ Volver a la sección anterior       │
├────────────────────────────────────┤
│ Footer                             │
└────────────────────────────────────┘
```

### Escritorio (≥ 1024 px)

```text
┌───────────────────────────────────────────────────────────┐
│ Header: logo | nombre del concurso | usuario | Salir      │
├───────────────────────────────────────────────────────────┤
│               │ Sección 3 de 10        [Guardar progreso] │ ← sticky
│ Secciones     │ Familiares                                │
│ ✓ Personal    │▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│ ✓ Domicilio   ├───────────────────────────────────────────┤
│ ● Familiares  │ Instrucciones y campos                    │
│ ○ Cargo       │                                           │
│ …             │ Último guardado: hoy, 14:32               │
│               │ Sección anterior     [Guardar y continuar]│
├───────────────┴───────────────────────────────────────────┤
│ Footer: ayuda | privacidad | soporte | bases del concurso │
└───────────────────────────────────────────────────────────┘
```

*(Los iconos de texto ✓ ● ○ son solo esquemáticos para el diagrama; en la interfaz se usan iconos de lucide según `design/tokens.md`).*

---

## 3. Reglas de Sticky y Scroll

- Solo la barra de sección es sticky; el header se desplaza con la página.
- `html { scroll-padding-top }` igual a la altura de la barra sticky, para que el foco nunca quede tapado por elementos flotantes (WCAG 2.4.11).
- Sin scroll horizontal a 320 px.
