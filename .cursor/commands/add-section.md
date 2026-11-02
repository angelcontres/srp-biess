<!-- Agrega un bloque de contenido o un grupo de campos a una página existente del sistema de postulación BIESS, respetando el espaciado, la escala tipográfica y el layout responsive de DESIGN.md. Úsala cuando se pida añadir un bloque, aviso, grupo de preguntas o lista a una pantalla ya creada. -->

# Agregar una sección a una página

Argumentos: el texto escrito después del comando `/add-section` = `"ruta de la página: tipo de sección"` (por ejemplo `"src/pages/Resumen.tsx: aviso institucional"`). Si no hay texto, toma estos datos del mensaje del usuario. Si falta la página o el tipo, pregunta antes de seguir.

Nota de vocabulario: en la interfaz, "Sección" es cada una de las 10 partes de la postulación. Aquí "sección" significa un **bloque dentro de una página**. No crees una nueva sección de postulación con esta skill; para eso existe `new-form-section`.

## 1. Lee las fuentes

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto. Si no existe, detente y avisa.
2. Lee la página indicada completa: su jerarquía de encabezados, contenedor, espaciado entre bloques y componentes usados.
3. Si el bloque tiene contenido institucional o legal, búscalo en `08-diseno-interfaz-usuario.md` o en el SDD. Si no está, usa `[PENDIENTE: descripción]`. No lo inventes.

## 2. Elige el patrón

| Tipo pedido | Patrón (DESIGN.md) |
|---|---|
| Bloque de contenido | `section` con `h2` (`text-lg font-semibold`) + texto `text-base max-w-prose`. |
| Aviso institucional | Alerta `info` (§7.6): `bg-info-50 border border-primary border-l-4 rounded-md p-4`, icono `Info`. |
| Grupo de campos | `fieldset` con `legend` (estilo `h2` o `h3`) y campos según §7.4, `space-y-6`. |
| Preguntas Sí/No | Un `fieldset` por pregunta, radios "Sí" / "No" sin preselección, campos condicionales justo debajo. Sin toggles. |
| Lista repetible | Tarjetas de ítem (§7.7), estado vacío de solo texto (§9), botón secundario "Agregar" al final. |
| Resumen o datos de solo lectura | Lista de definiciones (`dl`) o tarjeta con `border-line-strong`; valores con `tabular-nums` cuando son números o fechas. |
| Estado o historial | Lista ordenada cronológica inversa, fechas con `Intl` `es-EC`, badges de §7.8. |

Si el tipo pedido no encaja en ninguno, describe al usuario qué propones y espera confirmación.

## 3. Insértalo

- Ubícalo en el orden lógico de lectura. La posición en el DOM debe coincidir con la visual.
- **Espaciado**: usa el mismo `space-y-*` del contenedor padre (`space-y-8` o `space-y-12` entre bloques). No añadas márgenes sueltos que dupliquen el espacio.
- **Encabezados**: el nivel sigue al de los bloques vecinos. Nunca un segundo `h1` ni saltos de nivel.
- **Ancho**: dentro del contenedor existente (`max-w-5xl` o `max-w-7xl`). No abras contenedores de ancho completo.
- **Responsive**: una columna en < 640 px; dos columnas solo para campos cortos relacionados desde `sm:`; nada de anchos fijos en px.
- **Tokens**: solo clases de DESIGN.md §2 y §13; contenedores con `border-line-strong`, sin sombra.
- **Un solo botón primario** en la página: si el bloque necesita una acción, usa un botón secundario o terciario.
- Si la página es una sección del formulario de postulación y el bloque agrega campos:
  - Añádelos al esquema de validación.
  - Inclúyelos en el borrador local (`sessionStorage`) y en la carga y guardado de datos.
  - Conecta los errores al resumen de errores.
- No modifiques otros bloques salvo para mantener el espaciado consistente; si lo haces, dilo en el informe.

## 4. Verifica e informa

1. Ejecuta `lint`, `typecheck` y `build` si existen.
2. Comprueba que no introdujiste hex, valores arbitrarios, `ring-`, `rounded-xl` o mayores, `uppercase`, `·` ni emojis.
3. Informa: qué se insertó, dónde, qué patrón se usó y cualquier `[PENDIENTE]`.
