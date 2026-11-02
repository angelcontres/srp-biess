# Lo que Nunca Debe Aparecer y Checklist de Entrega

Este archivo consolida las prohibiciones estrictas de la interfaz y la lista de comprobación obligatoria antes de dar por terminada cualquier pantalla o componente.

---

## 1. Lo que nunca debe aparecer

| ❌ Nunca | ✅ En su lugar |
|---|---|
| Emojis | Iconos de Lucide con texto |
| Punto medio `·` como separador | Coma, salto de línea o columnas |
| `uppercase`, `tracking-widest` en etiquetas o títulos | Mayúscula inicial |
| `backdrop-blur`, glassmorphism, overlays difuminados | `bg-neutral-900/60` sólido |
| Degradados, shimmer, brillos | Colores sólidos de tokens |
| Sombra para delimitar tarjetas | `border border-line-strong` |
| `rounded-xl` o mayores, botones píldora | `rounded-md` / `rounded-lg` |
| `ring` para el foco | `outline` |
| Spinner a pantalla completa o centrado en la página | Skeleton (layout) o spinner en botón (operación) |
| Hero promocional, imágenes de stock, ilustraciones, carruseles | Encabezado institucional con texto sobrio |
| Animaciones de entrada por scroll, hover que mueve tarjetas | Sin movimiento |
| Grids de tarjetas de métricas estilo dashboard | Listas y tarjetas de contenido real |
| Placeholder como etiqueta | `label` visible |
| Toggles para Sí/No | Radio buttons |
| Opciones preseleccionadas en declaraciones o consentimientos | Ninguna marcada por defecto |
| Más de un botón primario por pantalla | Un primario + secundarios |
| `danger` para "no elegible" | Aviso institucional `info` |
| Truncar títulos o botones con `…` | Permitir que el texto pase a otra línea |
| Fuente monoespaciada para códigos o etiquetas | Inter con `tabular-nums` |
| Flechas `→` en botones y enlaces | Texto del verbo |
| Paletas por defecto de Tailwind | Tokens oficiales de este sistema |

---

## 2. Checklist antes de entregar una pantalla

- [ ] Solo clases de tokens; ningún valor hexadecimal suelto ni paleta por defecto de Tailwind.
- [ ] Todo contenedor con `border-line-strong`; inputs con `border-line-input`.
- [ ] Un solo botón primario por pantalla; botón "Guardar progreso" visible en la barra sticky.
- [ ] Foco con `outline`, visible y nunca tapado por la barra sticky (`scroll-padding-top`).
- [ ] Todo estado comunicado con icono + texto (nunca solo con color).
- [ ] Sin emojis, sin `·`, sin mayúsculas sostenidas, sin desenfoque, degradados ni shimmer.
- [ ] Sin scroll horizontal a 320 px; controles y áreas táctiles de 44 px de alto mínimo.
- [ ] Carga con skeleton para layout o spinner en botón para operaciones puntuales.
- [ ] Error de backend: los datos permanecen en los inputs y se puede reintentar sin recargar.
- [ ] Borrador local validado contra el esquema al restaurar y borrado al guardar o cerrar sesión.
- [ ] Textos con trato de usted, términos fijos oficiales y fechas formateadas con `Intl` `es-EC`.
- [ ] Auditoría axe sin incidencias críticas ni graves; recorrido completo operable con teclado.
