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
│ …             │                                           │
│───────────────│                                           │
│ Último        │                                           │
│ guardado:     │                                           │
│ hoy, 14:32    │ Sección anterior     [Guardar y continuar]│
│   ↑ sticky    │                                           │
├───────────────┴───────────────────────────────────────────┤
│ Footer: ayuda | privacidad | soporte | bases del concurso │
└───────────────────────────────────────────────────────────┘
```

*(Los iconos de texto ✓ ● ○ son solo esquemáticos para el diagrama; en la interfaz se usan iconos de lucide según `design/tokens.md`).*

---

## 3. Reglas de Sticky y Scroll

- Son sticky la barra de sección y, desde 1024 px, la columna de secciones (`lg:sticky lg:top-0 lg:self-start`, con `lg:max-h-screen lg:overflow-y-auto` si no cabe). El header se desplaza con la página.
- Desde 1024 px, el indicador de guardado va debajo de la lista de secciones, siempre a la vista; por debajo de 1024 px va al pie del formulario. Se dibuja una sola vez para no duplicar la región `role="status"`.
- Contenido más alto que la ventana dentro de un contenedor con acciones (modal): la cabecera y los botones quedan siempre visibles y solo el cuerpo se desplaza con su barra de scroll. Ver `design/components/modal.md`.
- `html { scroll-padding-top }` igual a la altura de la barra sticky, para que el foco nunca quede tapado por elementos flotantes (WCAG 2.4.11).
- Sin scroll horizontal a 320 px.
