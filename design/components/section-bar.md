# Barra de Sección con Línea de Progreso

Es el componente principal que orienta al ciudadano durante todo el trámite. Aparece en todas las secciones editables y en todos los tamaños de pantalla.

---

## Estructura

```tsx
<div className="sticky top-0 z-30 bg-surface border-b border-line-strong">
  <div className="mx-auto max-w-5xl px-4 py-3 flex items-start justify-between gap-3">
    <div className="min-w-0">
      <p className="text-sm text-muted tabular-nums">Sección 3 de 10</p>
      <h1 tabIndex={-1} className="text-xl sm:text-2xl font-bold leading-tight text-neutral-900">
        Familiares
      </h1>
    </div>
    <ButtonSecondary className="shrink-0 whitespace-nowrap">Guardar progreso</ButtonSecondary>
  </div>
  <div
    role="progressbar"
    aria-valuemin={0}
    aria-valuemax={10}
    aria-valuenow={5}
    aria-valuetext="5 de 10 secciones completas"
    className="h-1 w-full bg-line"
  >
    <div className="h-full bg-primary motion-safe:transition-[width] motion-safe:duration-200" style={{ width: '50%' }} />
  </div>
</div>
```

---

## Reglas de Implementación

- **Diferencia semántica:** "Sección N de 10" indica **la posición actual** del usuario. La línea de progreso indica las **secciones completadas según el servidor**. Pueden diferir, y es correcto que así sea.
- **Botón "Guardar progreso":** Queda a la **derecha**, siempre visible, con altura de 44 px (`h-11`), **nunca se encoge ni se abrevia**. Si falta espacio en móvil, el título pasa a dos líneas; nunca se trunca con puntos suspensivos (`…`).
- **Línea sólida:** `bg-primary` sobre `bg-line`, sin esquinas redondeadas, rayas, degradados ni efectos de brillo.
- **Manejo del foco y navegación:** Al cambiar de sección, el foco programático va al `h1` (`tabIndex={-1}`) y la etiqueta `<title>` cambia a *"Sección 3 de 10: Familiares | Postulación BIESS"*.
- **Ciclo de estados del botón:** "Guardar progreso", luego "Guardando…" con spinner integrado (`LoaderCircle`), y vuelve a "Guardar progreso". El resultado final se informa en el indicador de guardado (`design/components/save-status.md`) y con un toast (`design/components/toast.md`).
