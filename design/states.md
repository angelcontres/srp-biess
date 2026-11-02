# Movimiento y Estados de la Interfaz

Este archivo define las directrices de animación sutil, así como el tratamiento estricto de los estados de carga, vacío, validación, errores técnicos, ErrorBoundary y el borrador local en memoria.

---

## 1. Movimiento

| Elemento | Movimiento |
|---|---|
| Línea de progreso | `transition-[width] duration-200` al completar una sección. |
| Modal | Fundido de entrada de 150 ms, sin desplazamiento ni escala. |
| Ayudas y campos condicionales | Aparecen directamente, sin animación. |
| Hover y foco | Cambio instantáneo; el foco nunca se anima. |
| Spinner | `animate-spin`, **solo dentro de botones**. |
| Skeleton | Como mucho `animate-pulse` (opacidad). Sin shimmer. |
| Cambio de sección | Ninguno. |

Todo movimiento va con el prefijo `motion-safe:`. Con `prefers-reduced-motion: reduce` no se mueve nada; los estados siguen comunicándose con texto.

---

## 2. Estados: Carga, Vacío y Errores

### Carga

- **Operaciones concretas** (guardar, enviar, subir, ingresar): spinner dentro del botón + texto. Nunca un spinner a pantalla completa ni centrado en la página.
- **Carga de layout** (abrir una sección, lista de ítems, "Mi postulación"): skeleton.
  - Bloques `bg-line rounded-md` que imitan la forma real (etiqueta + input, tarjeta), sin shimmer.
  - Solo si la carga supera 1 s (evita parpadeos innecesarios).
  - Contenedor con `aria-busy="true"` y texto oculto visualmente: "Cargando sección".

### Vacío

Solo texto y acción, sin ilustración ni icono:

> Aún no ha registrado formación académica. Use "Agregar" para incluir la primera.

Si la sección admite responder "No", la pregunta Sí/No va antes de la lista.

### Error de validación

Error inline en el campo + resumen de errores (`design/components/error-summary.md`).

### Error de backend (HTTP 500 / 503)

La aplicación sigue funcionando y **los datos siguen en los inputs**.

1. Alerta de error en línea, justo debajo de la barra sticky, con `role="alert"`:  
   **"No se pudo guardar."** Sus datos siguen en pantalla. Vuelva a presionar "Guardar progreso". *Código de referencia: [id]*
2. El botón vuelve a su estado normal; se puede reintentar **sin recargar**.
3. El indicador de guardado pasa a "No se guardaron los cambios".
4. Nunca se limpia el formulario ni se redirige.

Sin conexión (`navigator.onLine === false`): advertencia persistente *"Sin conexión. Sus cambios no se pueden guardar hasta que se restablezca."*

### Error de frontend (ErrorBoundary)

Pantalla con la estructura del sistema (header y footer), sin ilustraciones:

- Título: "Ocurrió un error en la página"
- Cuerpo: "Recargue la página para continuar." Solo si existe un borrador, añadir: "Sus datos no guardados se conservan en este navegador."
- Botón primario "Recargar página" (`RotateCw`).
- **Nunca** se dispara un guardado al servidor desde el ErrorBoundary: el estado en memoria puede estar corrupto.

### Borrador local (draft cache)

1. Un hook con debounce de 2 s guarda los valores de la sección en `sessionStorage` mientras el usuario escribe.
2. Clave: `draft:{hashDeCédula}:{idSección}`. Nunca la cédula en claro.
3. Al montar la sección, si hay borrador, **se valida contra el esquema del formulario** (el mismo de la validación). Si no pasa, se descarta sin restaurar.
4. Si pasa, se restaura y se muestra una alerta institucional: *"Se recuperaron datos no guardados del [hora]. Revíselos y presione “Guardar progreso”."* con la acción "Descartar borrador". El indicador queda en "Hay cambios sin guardar".
5. Se borra al guardar con éxito, al cerrar sesión y cuando expira la sesión.
6. No incluye archivos (los PDF se suben al servidor en el momento).
7. La persistencia oficial ocurre **solo** al presionar "Guardar progreso" o "Guardar y continuar".
8. El borrador sobrevive a recargas y crashes, pero no al cierre de la pestaña. Es intencional: reduce la exposición de datos personales.

### Pantallas de error del sistema (404, servicio no disponible, fase cerrada)

Header y footer normales, título claro, una línea de explicación, qué hacer y un botón ("Volver al inicio", "Reintentar"), más el enlace a soporte. Sin ilustraciones, números gigantes ni humor.
