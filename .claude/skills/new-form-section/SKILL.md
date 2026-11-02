---
name: new-form-section
description: Crea una de las 10 secciones del formulario de postulación BIESS (Personal, Domicilio, Familiares, Cargo, Formación, Experiencia, Publicaciones, Capacitación, Inhabilidades, Resumen) con barra de sección sticky, guardado, borrador local, resumen de errores y estados de carga y error según DESIGN.md.
---

# Crear una sección del formulario de postulación

Argumentos: `$ARGUMENTS` = nombre de la sección (por ejemplo `Formación`). Si ves el texto `$ARGUMENTS` sin sustituir, toma el nombre del mensaje del usuario; si no hay ninguno, pregúntalo.

## 1. Lee las fuentes

1. Lee **`DESIGN.md`** completo, en la raíz del proyecto, en especial §4 (wireframes), §7.1 (barra de sección), §7.4 (campo), §7.5 (resumen de errores), §7.9 (adjunto), §7.12 (indicador de guardado) y §9 (estados, error de backend, ErrorBoundary, borrador local). Si no existe, detente y avisa.
2. Busca la sección en `08-diseno-interfaz-usuario.md` (inventario 8.4 y especificaciones 8.5) y en los documentos de reglas de negocio del SDD. **Los campos, preguntas, validaciones y textos salen de ahí.** Si no están definidos, pregunta al usuario; no inventes campos.
3. Identifica el número de la sección según este orden: 1 Personal, 2 Domicilio, 3 Familiares, 4 Cargo, 5 Formación, 6 Experiencia, 7 Publicaciones, 8 Capacitación, 9 Inhabilidades, 10 Resumen. Elegibilidad no es una sección: es la pantalla de resultado tras Cargo.
4. Revisa una sección ya implementada, si existe, y copia su estructura, hooks y servicios. La consistencia entre secciones es obligatoria.

## 2. Infraestructura compartida

Antes de crear la sección, busca estas piezas. Si existen, reutilízalas; si faltan, créalas una sola vez en una ubicación compartida, siguiendo DESIGN.md, y avisa en el informe:

| Pieza | Responsabilidad |
|---|---|
| `SectionBar` | Barra sticky: "Sección N de 10", `h1` con el título, botón "Guardar progreso" a la derecha y línea de progreso (`role="progressbar"`). |
| `useDraftCache` | Borrador en `sessionStorage` con debounce de 2 s, clave `draft:{hashDeCédula}:{idSección}`, validación por esquema al restaurar, borrado al guardar con éxito, al cerrar sesión y al expirar la sesión. Sin archivos. |
| `useSaveSection` | Guarda en el servidor solo al presionar "Guardar progreso" o "Guardar y continuar". En error 500/503: no limpia datos, deja reintentar sin recargar y expone el estado para el indicador. |
| `SaveStatus` | Indicador de guardado (§7.12) con `role="status"`. |
| `ErrorSummary` | Resumen de errores con enlaces a campos; recibe el foco. |
| Guarda de cambios sin guardar | Confirmación al salir con cambios: "Guardar y salir" / "Salir sin guardar" / "Seguir editando". |

## 3. Estructura de la sección

En este orden:

1. `SectionBar` con número, título y estado de guardado.
2. Alerta de error de backend (si la hay), justo debajo de la barra.
3. Alerta de borrador recuperado (si aplica) con la acción "Descartar borrador".
4. `ErrorSummary` (si hay errores tras intentar guardar).
5. Instrucciones de la sección (`text-neutral-700`, máximo 3 líneas) y avisos institucionales si los hay.
6. Contenido:
   - **Campos simples**: §7.4, `space-y-6`, con "(obligatorio)" en texto, `autocomplete` correcto, fechas en día/mes/año, cédula con `inputMode="numeric"`.
   - **Datos verificados**: solo lectura con badge "Verificado".
   - **Listas repetibles** (Familiares, Formación, Experiencia, Publicaciones, Capacitación): pregunta Sí/No previa si aplica, tarjetas de ítem, estado vacío de solo texto, botón secundario "Agregar", modal (pantalla completa en móvil) para agregar y editar, confirmación al eliminar nombrando el ítem.
   - **Adjuntos PDF**: los cinco estados de §7.9.
   - **Inhabilidades**: texto jurídico literal, un `fieldset` por pregunta, radios sin preselección, campos condicionales justo debajo.
   - **Cargo**: tarjetas seleccionables (§7.7), aviso de selección definitiva y confirmación en modal; al confirmar, la navegación va a la pantalla de Elegibilidad.
   - **Resumen**: estado de cada sección con badges, datos en solo lectura, consentimientos sin preselección, botón primario "Enviar postulación" y confirmación en modal con foco inicial en "Seguir revisando".
7. Pie: `SaveStatus`, botón primario "Guardar y continuar" (en Resumen: "Enviar postulación") y enlace terciario "Volver a la sección anterior". En < 640 px, apilados a ancho completo.

## 4. Comportamiento obligatorio

- Al montar: skeleton si la carga supera 1 s (`aria-busy`), foco en el `h1`, `<title>` "Sección N de 10: [Título] | Postulación BIESS".
- Validar al salir del campo y al guardar; nunca mientras se escribe.
- "Guardar y continuar" no avanza si hay errores bloqueantes: muestra el resumen.
- La línea de progreso refleja secciones completadas **según el servidor**, no las visitadas.
- Secciones bloqueadas por elegibilidad se muestran "No disponible" en el stepper, no se ocultan.
- Ningún guardado al servidor desde el ErrorBoundary.

## 5. Verifica e informa

1. Ejecuta `lint`, `typecheck`, `test` y `build` si existen.
2. Repasa DESIGN.md §14 punto por punto.
3. Informa: archivos creados, piezas compartidas reutilizadas o creadas, campos implementados con su fuente en el SDD y cualquier `[PENDIENTE]` (por ejemplo, textos jurídicos).
