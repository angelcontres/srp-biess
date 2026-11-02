<!-- Crea una nueva página o ruta del sistema de postulación BIESS siguiendo el layout, la tipografía y el espaciado de DESIGN.md. Úsala cuando se pida una pantalla nueva (login, resultado, Mi postulación, error, landing). Para secciones del formulario de postulación usa new-form-section. -->

# Crear una nueva página

Argumentos: el texto escrito después del comando `/new-page` = nombre o ruta de la página (por ejemplo `mi-postulacion` o `/recuperacion`). Si no hay texto, toma el nombre del mensaje del usuario; si no hay ninguno, pregúntalo antes de seguir.

## 1. Lee las fuentes antes de escribir código

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto. No asumas que lo conoces. Presta atención a §0 (decisiones), §4 (layout y breakpoints), §5 (forma), §7 (componentes), §9 (estados), §12 (prohibiciones) y §13 (configuración de Tailwind). Si `DESIGN.md` no existe, detente y avisa al usuario.
2. Si existe `08-diseno-interfaz-usuario.md` (u otro documento del SDD), busca la pantalla en su inventario (8.4) y su especificación (8.5). El contenido y las reglas de negocio salen de ahí, no de tu criterio.
3. Revisa `package.json` para confirmar el stack: React, Tailwind 3.4.17, `lucide-react` y el enrutador (React Router, Next.js u otro).
4. Localiza cómo están organizadas las rutas, las páginas y los componentes existentes (`src/pages`, `src/routes`, `app/`, `src/components`…). Sigue esa convención. Si no puedes determinarla, pregunta en lugar de inventar una estructura.

## 2. Clasifica la página

| Tipo | Ejemplos | Layout (DESIGN.md) |
|---|---|---|
| Acceso | Login, registro, recuperación, restablecer, bloqueo | Tarjeta centrada `max-w-md w-full bg-surface border border-line-strong rounded-lg p-6 sm:p-8` sobre `bg-background`. Sin sombra. |
| Resultado o estado | Elegibilidad, postulación enviada, Mi postulación, fase cerrada | Contenido `max-w-5xl mx-auto px-4`, `h1` de página, bloques con `space-y-12`. |
| Error del sistema | 404, servicio no disponible | Header y footer normales, título, una línea de explicación, botón y enlace a soporte. Sin ilustraciones. |
| Landing | Página pública de inicio | `max-w-7xl mx-auto px-4`, `h1` con escala `display` (`text-3xl font-bold`). Encabezado institucional con texto, sin hero promocional. |
| Sección del formulario | Personal, Domicilio, Familiares… | **No uses esta skill.** Usa `new-form-section`. |

## 3. Construye la página

- Reutiliza los componentes que ya existan: header, footer, franja `secondary`, enlace "Saltar al contenido", botones, alertas, badges, tarjetas. Búscalos antes de escribir JSX nuevo. Si falta uno que DESIGN.md §7 especifica, créalo siguiendo esa especificación (o usa la skill `new-component`). **No inventes componentes que DESIGN.md no describe**; si crees que hace falta uno, pregunta.
- Header y footer en todas las páginas, con landmarks `header`, `main`, `footer`.
- Un solo `h1` por página, jerarquía sin saltos (`h1` → `h2` → `h3`).
- `<title>` único: "[Nombre de la página] | Postulación BIESS".
- Ancho: `max-w-5xl` (formularios y estados) o `max-w-7xl` (landing), siempre `mx-auto px-4`.
- Separación entre bloques: `space-y-8` o `space-y-12` (32–48 px). Entre campos: `space-y-6`.
- Solo clases de tokens (DESIGN.md §2 y §13). Ningún hex, ningún valor arbitrario (`p-[13px]`), ninguna paleta por defecto (`gray-*`, `slate-*`, `blue-*`…). Las paletas por defecto no existen en la configuración: si las usas, el estilo desaparece en silencio.
- Contenedores con `border border-line-strong`; inputs con `border-line-input`; divisores internos con `border-line`.
- Radios: botones e inputs `rounded-md`, tarjetas y modales `rounded-lg`, badges `rounded`. Nada mayor que `rounded-lg`.
- Un único botón primario por pantalla.
- Foco con `outline` (nunca `ring`).
- Iconos de `lucide-react` según la tabla de DESIGN.md §6, siempre con texto y `aria-hidden="true"`.
- Carga de datos: skeleton (DESIGN.md §9) si supera 1 s. Spinner solo dentro de botones.
- Errores de servidor: alerta en línea, sin limpiar datos ni redirigir (DESIGN.md §9).
- Textos en usted, términos fijos ("postulación", "sección"), fechas con `Intl` `es-EC` y zona `America/Guayaquil`.
- Contenido institucional o legal que no tengas: escribe `[PENDIENTE: descripción]` visible. No lo inventes.

## 4. Registra la ruta

- Agrégala al enrutador siguiendo el patrón existente.
- Si es una página autenticada, usa la misma protección de ruta que las demás. No crees un mecanismo nuevo.

## 5. Verifica

1. Ejecuta los scripts de verificación que tenga el proyecto (`lint`, `typecheck`, `build`, `test`). Corrige lo que falle.
2. Recorre la lista de DESIGN.md §14 y confirma cada punto en el código que escribiste.
3. Busca en tus archivos nuevos: `#[0-9a-fA-F]{3,8}`, `\[[0-9.]+px\]`, `ring-`, `rounded-(xl|2xl|3xl)`, `backdrop-blur`, `uppercase`, `·`, emojis. No debe aparecer ninguno.

## 6. Informa

En pocas líneas: archivos creados o modificados, ruta registrada, componentes reutilizados y cualquier `[PENDIENTE]` que haya quedado.
