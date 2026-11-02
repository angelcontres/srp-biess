# Stepper (Lista de Secciones)

Componente de navegación que refleja las secciones del formulario de postulación y su estado respectivo.

---

## Comportamiento Responsive

- **≥ 1024 px:** Barra lateral izquierda, elemento semántico `<nav aria-label="Secciones de la postulación">`, lista ordenada (`<ol>`).
- **640–1023 px:** Franja horizontal bajo el header con desplazamiento horizontal propio (no de la página entera).
- **< 640 px:** No se muestra de forma permanente. Un enlace interactivo *"Ver todas las secciones"* dentro del contenido principal abre la lista accesible.

---

## Estados de los Ítems

| Estado | Icono / Marca | Texto visible | Texto para lector de pantalla |
|---|---|---|---|
| Completada | `CircleCheck` `text-success` | nombre | `, completada` |
| Actual | número en círculo `bg-primary text-white` | nombre en `font-semibold` | `aria-current="step"` |
| Pendiente | número en círculo `border-line-input` | nombre | `, pendiente` |
| Con errores | `CircleAlert` `text-danger` | nombre + "Con errores" | — |
| No disponible | `Lock` `text-muted` | nombre + "No disponible" | — |

---

## Reglas

- Cada ítem interactivo tiene un área táctil de **44 px de alto como mínimo**.
- Los ítems en estado "No disponible" no deben ser enlaces ni permitir eventos de foco o navegación.
