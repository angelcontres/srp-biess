---
name: responsive-check
description: Verifica que una página del sistema de postulación BIESS funcione en todos los breakpoints de DESIGN.md (320 a 1280 px y zoom), detectando desbordes, superposiciones, controles pequeños y foco tapado, y propone arreglos con tokens existentes. Úsala tras crear o cambiar una pantalla.
---

# Revisión responsive

Argumentos: `$ARGUMENTS` = ruta de la página (archivo o ruta de la aplicación). Si ves el texto `$ARGUMENTS` sin sustituir, toma la página del mensaje del usuario; si no hay ninguna, pregúntala.

**Esta skill reporta y sugiere. No aplica cambios sin confirmación del usuario.**

## 1. Prepara

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto, en especial §4 (breakpoints y wireframes), §5, §7.1 (barra de sección), §7.10 (modal) y §10. Si no existe, detente y avisa.
2. Lee el archivo de la página y los componentes que importa.

## 2. Anchos a revisar

| Ancho | Representa | Comportamiento esperado (DESIGN.md §4) |
|---|---|---|
| 320 px | Mínimo de reflujo (WCAG 1.4.10) | Una columna, sin scroll horizontal. |
| 360 px | Android de gama baja | Igual; "Guardar progreso" sin encogerse. |
| 640 px | `sm:` | Dos columnas solo en campos cortos relacionados. |
| 768 px | Tablet vertical | Stepper horizontal bajo el header. |
| 1024 px | `lg:` | Stepper lateral, contenido `max-w-5xl`. |
| 1280 px | Escritorio | Contenido centrado, sin estirarse. |
| 1280 px con zoom 200 % | Texto ampliado | Sin pérdida de contenido. |

## 3. Método

**A. Revisión estática (siempre)**

Busca en el código de la página y sus componentes:
- Anchos o altos fijos: `w-[`, `min-w-[`, `h-[`, `style={{ width:` con px.
- Contenedores sin `min-w-0` dentro de `flex` (provocan desborde de texto largo).
- `whitespace-nowrap` o `truncate` en títulos o etiquetas (solo se permite `whitespace-nowrap` en el botón "Guardar progreso"; nunca truncar).
- Grids con columnas fijas sin variante móvil (`grid-cols-3` sin `grid-cols-1` base).
- Tablas sin alternativa en tarjetas para < 640 px.
- Modales sin modo pantalla completa en < 640 px.
- Controles con alto menor de 44 px (`h-8`, `h-9`, `h-10`) o separación menor de 8 px entre controles adyacentes.
- Elementos `fixed` o `sticky` distintos de la barra de sección.
- Falta de `scroll-padding-top` acorde a la altura de la barra sticky.
- Imágenes o SVG sin `max-w-full`.

**B. Revisión en navegador (si es posible)**

Si el proyecto puede levantarse (`npm run dev` o similar) y tienes un navegador automatizado disponible (Playwright, Puppeteer o el navegador del agente):
1. Abre la página en cada ancho de la tabla.
2. Comprueba `document.documentElement.scrollWidth > window.innerWidth` (scroll horizontal).
3. Recorre la página con Tab y verifica que el elemento enfocado no quede bajo la barra sticky.
4. Captura pantallas de 360, 768 y 1280 px para el informe.

Si no puedes ejecutarla, dilo y entrega solo la revisión estática.

## 4. Comprobaciones específicas del proyecto

- **Barra de sección**: "Sección N de 10" y el título a la izquierda, "Guardar progreso" a la derecha con 44 px; el título pasa a dos líneas antes que encoger el botón; la línea de progreso ocupa el ancho completo.
- **Botones del final**: `w-full` apilados en < 640 px (primario arriba); en fila desde `sm:`.
- **Tarjetas de cargo**: apiladas en < 1024 px, nunca tres columnas estrechas.
- **Stepper**: oculto en < 640 px (enlace "Ver todas las secciones"); horizontal con desplazamiento propio en 640–1023 px; lateral en ≥ 1024 px.
- **Teclado virtual**: el campo enfocado y su error no deben quedar ocultos.
- **Toasts**: abajo al centro en móvil, sin tapar los botones del final.

## 5. Reporta

| Ancho | Elemento | Problema | Arreglo sugerido (tokens existentes) |
|---|---|---|---|
| 360 px | Botón "Guardar progreso" | Se encoge por `flex-1` | `shrink-0 whitespace-nowrap`; permitir que el título haga salto de línea con `min-w-0` |

Al final: resumen de anchos sin problemas, problemas encontrados por gravedad y la pregunta de si el usuario quiere que apliques los arreglos.
