# Indicador de Guardado (Save Status)

Elemento accesible que informa sobre la sincronización con la base de datos:

- **≥ 1024 px:** debajo de la lista de secciones, en la columna lateral sticky, para que esté siempre a la vista.
- **< 1024 px:** junto a los botones de navegación, al pie del formulario.

Se renderiza una sola vez según el ancho (nunca dos regiones `role="status"` a la vez).

---

## Estructura Base

Clases: `text-sm text-muted flex items-center gap-1.5`  
Atributo accesible: `role="status"`

---

## Estados del Indicador

| Estado | Texto Visible | Icono (size=16) |
|---|---|---|
| **Sin cambios** | "Último guardado: hoy, 14:32" | `CircleCheck` |
| **Cambios pendientes** | "Hay cambios sin guardar" | `TriangleAlert` (`text-warning`) |
| **Guardando** | "Guardando…" | `LoaderCircle` (`animate-spin`) |
| **Error** | "No se guardaron los cambios" | `CircleAlert` (`text-danger`) |

---

## Formato de Fecha

El texto de la hora y fecha debe generarse siempre utilizando la API nativa de JavaScript `Intl.DateTimeFormat` configurada con el locale `es-EC` y la zona horaria `America/Guayaquil` (ver `design/content.md`).
