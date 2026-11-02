# 08. Diseño de interfaz de usuario — SDD

## 8.0 Alcance y convenciones

Este documento especifica **cómo se presenta** el servicio de postulación al ciudadano. No define reglas de negocio, plazos, políticas de seguridad ni textos jurídicos; cuando la interfaz depende de ellos, se indica la referencia y el requisito de validación.

| Marca | Significado |
|---|---|
| **[O]** | Obligatorio. Su incumplimiento bloquea la aceptación de la pantalla. |
| **[R]** | Recomendado. Se aplica salvo justificación técnica documentada. |
| **[V]** | Requiere validación funcional, institucional o jurídica antes de implementarse en su forma final. |

Referencias cruzadas (no se duplican aquí):

| Tema | Documento responsable |
|---|---|
| Duración de bloqueo, política de contraseñas, duración de sesión, intentos permitidos | Requisitos de seguridad del SDD. |
| Reglas de elegibilidad, nacionalidad, selección de cargo, inhabilidades | Reglas de negocio del SDD. |
| Estados de la postulación y sus transiciones | Modelo de dominio / estados del SDD. |
| Límites de archivos (tipo, tamaño, cantidad) | Reglas de negocio / almacenamiento del SDD. |
| Textos legales de consentimientos, declaraciones e inhabilidades | Área jurídica del BIESS. |
| Contenido y formato del PDF del formulario | Especificación de reportes/documentos del SDD. |

Norma de accesibilidad de referencia: **WCAG 2.2 nivel AA**.

---

## 8.1 Principios de diseño

1. **Claridad:** el postulante siempre sabe en qué sección está, qué falta, qué debe hacer y qué ocurrirá después.
2. **Progresividad:** formulario largo dividido en secciones con guardado independiente; un tema por pantalla.
3. **Confianza:** mensajes institucionales, datos verificados identificados como tales y confirmaciones antes de acciones irreversibles.
4. **Transparencia:** los requisitos se explican antes de solicitar información; las consecuencias de cada acción importante se informan antes de ejecutarla.
5. **Neutralidad institucional:** apariencia sobria de servicio público. Sin recursos promocionales, imágenes de stock, carruseles, ilustraciones decorativas ni lenguaje comercial.
6. **Accesibilidad:** WCAG 2.2 AA; usable con teclado, lector de pantalla, zoom al 200 % y por personas con baja alfabetización digital.
7. **Responsive y tolerante a mala conectividad:** usable en móvil de gama baja y redes lentas sin pérdida de información.
8. **Consistencia:** mismos patrones de campos, botones, alertas, estados y vocabulario en todas las pantallas.

### 8.1.1 Tres capas de información (diferenciación visual) [O]

| Capa | Contenido | Tratamiento |
|---|---|---|
| **Institucional** | Normas, resultados administrativos, avisos de fase, textos legales. | Bloque con borde izquierdo `primary` de 4 px, fondo `info-50`, icono de información y título. No usa colores de error. |
| **Instrucciones** | Qué se pide y cómo completarlo. | Texto de cuerpo `neutral-700`, encima de los campos, máximo 3 líneas por bloque. |
| **Acciones del usuario** | Campos, botones, enlaces. | Componentes interactivos del sistema con estados definidos. |

### 8.1.2 Lenguaje [O]

- Tratamiento de **usted** en todo el sistema (sin mezclar con tú).
- Oraciones cortas (≤ 25 palabras), voz activa, vocabulario común. Términos técnicos o legales (p. ej., “inhabilidad”, “código dactilar”) llevan ayuda contextual.
- Los botones nombran la acción concreta (“Guardar y continuar”, “Enviar postulación”), nunca “Aceptar” u “OK” en acciones con consecuencias.
- Una acción conserva el mismo nombre en botón, confirmación y mensaje resultante (botón “Enviar postulación” → mensaje “Postulación enviada”).

---

## 8.2 Sistema de diseño

### Paleta

Valores existentes conservados. Contrastes medidos sobre blanco (`#FFFFFF`) y `neutral-50`.

| Token | Valor | Uso | Contraste / restricción |
|---|---|---|---|
| `primary` | `#0B5CAD` | Acciones principales, header, enlaces. | 6.67:1 sobre blanco. Apto para texto. |
| `primary-dark` | `#083F78` | Hover, énfasis, anillo de foco. | 10.54:1. Apto para texto. |
| `secondary` | `#F2B705` | Acentos institucionales. | **1.82:1 sobre blanco: prohibido como color de texto, borde de campo o único indicador de estado [O].** Solo como acento no informativo o con texto `neutral-900` encima (9.82:1). Sobre `primary` alcanza 3.67:1 (válido para indicadores no textuales). |
| `success` | `#1E7B45` | Confirmaciones. | 5.28:1. |
| `warning` | `#B45309` | Advertencias. | 5.02:1. |
| `danger` | `#B42318` | Errores de validación y fallos técnicos. | 6.57:1. **No se usa para resultados administrativos** (ver 8.7). |
| `neutral-50` | `#F8FAFC` | Fondo de página. | — |
| `neutral-900` | `#0F172A` | Texto principal. | 17.06:1 sobre `neutral-50`. |

Tokens complementarios necesarios (la paleta original no define grises intermedios ni fondos de alerta, imprescindibles para bordes de campo, texto secundario y alertas legibles). Valores propuestos, verificados:

| Token | Valor | Uso | Contraste |
|---|---|---|---|
| `neutral-0` | `#FFFFFF` | Superficie de tarjetas, campos y modales. | — |
| `neutral-100` | `#F1F5F9` | Fondo de campos de solo lectura. | Texto `neutral-600`: 6.92:1. |
| `neutral-200` | `#E2E8F0` | Divisores y bordes **decorativos** de tarjetas. | 1.23:1: no apto para bordes de campos. |
| `neutral-500` | `#64748B` | **Borde de campos**, iconos, paso pendiente del stepper. | 4.76:1 sobre blanco (≥ 3:1 exigido para componentes). |
| `neutral-600` | `#475569` | Texto secundario, ayudas. | 7.58:1. |
| `neutral-700` | `#334155` | Instrucciones. | > 10:1. |
| `info-50` | `#EFF6FF` | Fondo de avisos institucionales. | `primary` sobre él: 6.13:1. |
| `success-50` | `#ECFDF3` | Fondo de alerta de éxito. | `success`: 5.01:1. |
| `warning-50` | `#FFFBEB` | Fondo de advertencia. | `warning`: 4.84:1. |
| `danger-50` | `#FEF3F2` | Fondo de alerta de error. | `danger`: 6.05:1. |

Reglas:

- [O] Texto normal ≥ 4.5:1; texto grande (≥ 24 px o ≥ 18.66 px en negrita) ≥ 3:1; bordes de campos, iconos informativos y anillo de foco ≥ 3:1 frente a su fondo (WCAG 1.4.3 y 1.4.11).
- [O] Ningún valor de color se escribe directamente en componentes; solo tokens.
- [O] Todo estado comunicado por color se acompaña de texto y/o icono (WCAG 1.4.1).
- [R] Sin degradados, texturas ni fotografías decorativas.
- No se requiere modo oscuro. [O] La interfaz debe seguir siendo operable en modo de alto contraste del sistema (`forced-colors`): bordes y foco visibles, iconos no dependientes de color de fondo.

### Tipografía

- Familia: `Inter`, `system-ui`, sans-serif (se conserva).
- [R] `Inter` alojada en el propio servidor, en formato WOFF2 y subconjunto latino, con `font-display: swap`; si no carga, la interfaz debe ser plenamente legible con la fuente de sistema.
- Títulos: 600–700. Cuerpo: 400–500.
- Tamaños: `text-sm` (14 px), `text-base` (16 px), `text-lg` (18 px), `text-2xl` (24 px).

| Uso | Tamaño | Peso | Interlineado |
|---|---|---|---|
| Título de página (`h1`) | `text-2xl` | 700 | 1.25 |
| Título de bloque (`h2`) | `text-lg` | 600 | 1.35 |
| Cuerpo, etiquetas, **campos**, textos legales | `text-base` | 400–500 | 1.5 |
| Ayudas, metadatos (fechas de guardado, tamaño de archivo) | `text-sm` | 400 | 1.5 |

- [O] Campos de formulario y textos legales a 16 px como mínimo (además evita el zoom automático en iOS).
- [O] `text-sm` no se usa para instrucciones ni mensajes de error.
- [O] Ancho de línea del texto corrido ≤ 75 caracteres.
- [O] Códigos, cédulas y fechas con cifras tabulares (`font-variant-numeric: tabular-nums`).
- [O] Sin texto en mayúsculas sostenidas para etiquetas o párrafos; usar mayúscula inicial.
- [O] La página no pierde contenido con espaciado de texto ampliado (WCAG 1.4.12).

### Espaciado

- Escala Tailwind: 2, 4, 6, 8, 12, 16 (se conserva).
- Contenedor máximo: `max-w-5xl` para formularios; `max-w-7xl` para landing.
- Radios: `rounded-lg` (8 px) para campos y botones; `rounded-xl` (12 px) para tarjetas y modales.
- [O] Tarjetas: borde `neutral-200` de 1 px; sombra solo en elementos superpuestos.

| Elevación | Uso | Tratamiento |
|---|---|---|
| 0 | Página, secciones. | Sin sombra. |
| 1 | Tarjetas de ítems y de cargos. | Borde 1 px, sombra opcional muy suave. |
| 2 | Modales, menús, barra de acciones fija. | Sombra visible + overlay en modales. |

### Componentes base

Se conservan: botón primario, secundario, terciario y peligro; input, select, textarea, date, radio, checkbox; card, alert, badge, modal, tooltip, toast; stepper, tabs, breadcrumb; skeleton y spinner.

Se precisan como componentes del sistema (justificados por los flujos existentes, no son funcionalidades nuevas):

- **Campo de documento adjunto** (8.6.3).
- **Indicador de guardado** (8.6.2).
- **Resumen de errores** (8.6).
- **Bloque de aviso institucional** (8.1.1).
- **Badge de estado** con icono + texto.

Estados obligatorios por componente interactivo [O]:

| Estado | Requisito |
|---|---|
| Predeterminado | Borde/forma visible ≥ 3:1. |
| Hover | Cambio perceptible; no es el único indicador de interactividad. |
| Foco | Anillo de 2 px `primary-dark` con separación de 2 px; sobre fondo `primary`, anillo blanco. Nunca `outline: none` sin reemplazo. |
| Activo/presionado | Cambio de tono. |
| Deshabilitado | Atenuado + `disabled`/`aria-disabled`; la causa se explica en texto cercano. |
| Cargando | Spinner + texto (“Guardando…”); el control no acepta un segundo envío. |
| Error | Borde `danger` + icono + mensaje (8.6). |
| Solo lectura / verificado | Fondo `neutral-100`, sin borde de edición, etiqueta “Verificado” con icono. |

Uso de botones [O]:

| Variante | Uso | Regla |
|---|---|---|
| Primario | La acción principal de la pantalla. | **Uno solo por pantalla o modal.** |
| Secundario | Acciones alternativas (Guardar progreso, Volver). | — |
| Terciario (enlace) | Navegación, “Ver documento”. | — |
| Peligro | Eliminar ítems. | Siempre precedido de confirmación. |

Tooltip [O]: solo para ayudas breves no esenciales; se abre con foco y hover, se cierra con `Esc`, y no contiene información imprescindible ni enlaces (en móvil, usar ayuda desplegable en línea).

Movimiento [O]: sin animaciones decorativas. Se permiten transiciones ≤ 200 ms que muestran un cambio provocado por el usuario (abrir modal, desplegar ayuda). Con `prefers-reduced-motion: reduce`, se eliminan.

---

## 8.3 Layout y navegación

### Escritorio (> 1024 px)

```text
┌────────────────────────────────────────────────────────────┐
│ Header: logo BIESS | nombre del concurso | usuario | Salir │
├──────────────────┬─────────────────────────────────────────┤
│ Progreso         │ Paso 3 de 10 · Familiares               │
│ ✓ 1 Personal     │ Instrucciones de la sección             │
│ ✓ 2 Domicilio    │                                         │
│ ● 3 Familiares   │ Contenido de la sección                 │
│ ○ 4 Cargo        │                                         │
│ ○ …              │                                         │
│ ○ 10 Resumen     │ Último guardado: 14:32                  │
│                  │ [Guardar progreso]  [Guardar y continuar]│
├──────────────────┴─────────────────────────────────────────┤
│ Footer: ayuda | privacidad | soporte | bases del concurso  │
└────────────────────────────────────────────────────────────┘
```

- Tablet (640–1024 px): el stepper pasa a una franja horizontal bajo el header.
- Móvil (< 640 px): barra de progreso compacta (“Paso 3 de 10 · Familiares”) con botón “Ver todos los pasos” que despliega la lista.

> Se corrige la inconsistencia del documento original, que mostraba el stepper horizontal en el wireframe y una barra lateral en 8.9. La versión vigente es la de esta sección.

### Pasos del stepper

Los 10 pasos son: Personal, Domicilio, Familiares, Cargo, Formación, Experiencia, Publicaciones, Capacitación, Inhabilidades, Resumen. **Elegibilidad no es un paso**: es la pantalla de resultado que aparece tras confirmar el cargo. **[V]** Confirmar este orden con reglas de negocio.

### Navegación

- El stepper muestra secciones completadas, actual, pendientes y con errores.
- Se puede volver a una sección anterior sin perder datos.
- No se puede avanzar si la sección actual tiene errores bloqueantes; el resumen de errores indica cuáles.
- El resumen final muestra el estado de cada sección.
- [O] Las secciones posteriores a Cargo solo son accesibles cuando el resultado de elegibilidad lo permite; las bloqueadas se muestran como “No disponible” con el motivo, no se ocultan.
- [O] Las secciones bloqueadas por la regla de cargo definitivo no ofrecen la acción de cambiar cargo.
- [O] Al cambiar de paso: actualizar el `<title>` (“Paso 3 de 10: Familiares — Postulación BIESS”), llevar el foco al `h1` y conservar la posición de desplazamiento en la parte superior.
- [O] Si hay cambios sin guardar y el usuario intenta salir (otro paso, cerrar pestaña, cerrar sesión), se muestra una confirmación: “Tiene cambios sin guardar en esta sección.” con acciones “Guardar y salir” / “Salir sin guardar” / “Seguir editando”.
- [O] Enlace “Saltar al contenido” como primer elemento enfocable.
- [O] Landmarks: `header`, `nav` (stepper), `main`, `footer`.
- [O] Los enlaces de ayuda y soporte están en la misma ubicación en todas las pantallas (WCAG 3.2.6).
- [O] Los botones de acción no fijos aparecen al final del formulario; si se usa una barra de acciones fija en móvil, no debe tapar el campo enfocado (WCAG 2.4.11) y se compensa con `scroll-padding-bottom`.

---

## 8.4 Inventario de pantallas

| Pantalla | Elementos clave | Acciones |
|---|---|---|
| **Landing** | Título, descripción, requisitos, fase vigente y fechas, documentos a preparar, botones. | Iniciar sesión, registrarse. |
| **Login** | Cédula, contraseña, recuperación. | Ingresar. |
| **Registro** | Cédula, tipo de validación, valor, correo, contraseña, aviso de tratamiento de datos. | Validar, registrar. |
| **Bloqueo** | Mensaje, motivo, hora de desbloqueo, cuenta regresiva. | Reintentar al expirar. |
| **Recuperación** | Cédula, validación. | Enviar enlace. |
| **Restablecer** | Nueva contraseña, confirmación. | Guardar. |
| **Personal** | Datos verificados, estado civil, pareja. | Guardar progreso, guardar y continuar. |
| **Domicilio** | País, provincia, cantón, dirección, contacto. | Guardar progreso, guardar y continuar. |
| **Familiares** | Sí/No, lista, modal de agregar. | Agregar, editar, eliminar. |
| **Cargo** | Tres tarjetas de cargo, confirmación. | Seleccionar, confirmar. |
| **Elegibilidad** | Resultado, fundamento, fecha y hora, mensaje institucional. | Continuar o cerrar. |
| **Formación** | Lista, modal, PDF. | Agregar, editar, eliminar. |
| **Experiencia** | Lista, fechas, tiempo calculado, PDF. | Agregar, editar, eliminar. |
| **Publicaciones** | Sí/No, lista, PDF. | Agregar, editar, eliminar. |
| **Capacitación** | Lista, PDF. | Agregar, editar, eliminar. |
| **Inhabilidades** | 13 preguntas, campos condicionales. | Guardar progreso, guardar y continuar. |
| **Resumen** | Estado de secciones, revisión de datos, consentimientos. | Enviar postulación. |
| **Enviada** | Código, fecha y hora, estado, PDF. | Descargar, ir a Mi postulación. |
| **Mi postulación** | Estado, código, fechas, historial de estados, PDF. | Descargar. |
| **Sesión por expirar / expirada** | Aviso, tiempo restante. | Continuar sesión, iniciar sesión. |
| **Error / servicio no disponible** | Mensaje, qué hacer, contacto de soporte. | Reintentar, volver. |
| **Fase cerrada** | Aviso institucional, fechas. | Ir a Mi postulación (si existe). |

> **[V]** Botón de envío: el documento original lo llamaba “Guardar postulación”, pero la acción es un envío definitivo. Se adopta “Enviar postulación” para que la etiqueta describa la consecuencia. Validar con el área funcional.

---

## 8.5 Especificación de pantallas clave

### 8.5.1 Landing

- Encabezado institucional con título “Concurso Público de Méritos y Oposición — Directorio del BIESS” (no un *hero* promocional: sin imagen de fondo, sin eslogan).
- Texto breve del proceso: pasos en el orden real (registrarse, completar, revisar, enviar, descargar e imprimir y firmar).
- [O] Requisitos y documentos a preparar **antes** de registrarse: formato PDF, tamaño máximo, tipos de respaldo. Contenido según bases **[V]**.
- [O] Fase vigente y fechas de apertura y cierre, obtenidas de la configuración del sistema (no texto fijo).
- Botones “Iniciar sesión” y “Registrarse”.
- Aviso si la fase está cerrada, con bloque institucional y fechas; los botones de registro no se muestran como activos.
- Enlaces a bases, privacidad y soporte.
- [R] Tiempo estimado para completar la postulación, solo si existe una cifra validada.

### 8.5.2 Login

- Campo cédula (10 dígitos): `inputmode="numeric"`, `autocomplete="username"`, no `type="number"`.
- Campo contraseña con mostrar/ocultar (botón con `aria-pressed` y etiqueta “Mostrar contraseña”), `autocomplete="current-password"`.
- [O] Se permite pegar en todos los campos y el uso de gestores de contraseñas (WCAG 3.3.8).
- Enlace “¿Olvidó su contraseña?” (se corrige el tuteo).
- Botón “Ingresar”.
- Mensajes de error genéricos que no revelan si la cédula existe: “Cédula o contraseña incorrectas.”
- Estado de carga en el botón.
- [O] Si quedan pocos intentos antes del bloqueo, se informa antes de que ocurra (número según requisitos de seguridad **[V]**).

### 8.5.3 Registro

- Paso 1: cédula + tipo de validación (fecha de expedición o código dactilar, como radio buttons) + valor.
- [O] Ayuda contextual que indica dónde se encuentra el dato en la cédula. **[V]** Ubicación y formato según el modelo vigente de cédula.
- **[V]** El código dactilar es alfanumérico en la cédula ecuatoriana; el microcopy original “10 dígitos” puede ser incorrecto. Validar formato y ajustar validación y texto (8.10).
- [O] Fecha de expedición en tres campos (día, mes, año) o un campo de texto con formato `dd/mm/aaaa` visible; no depender solo del selector nativo de fecha.
- [O] Aviso de tratamiento de datos personales **antes** del Paso 1: qué datos se consultan, para qué y enlace a la política de privacidad. Texto **[V]** jurídico.
- Paso 2: correo + contraseña + confirmación (`autocomplete="new-password"`).
- Indicador de fortaleza de contraseña que muestra en texto cada requisito cumplido/pendiente (no solo una barra de color).
- Mensaje de bloqueo con cuenta regresiva.
- Al éxito, redirige al login con mensaje “Cuenta creada. Ingrese con su cédula y contraseña.”
- [O] Los fallos de validación de identidad distinguen entre: dato no coincide (validación), servicio de verificación no disponible (técnico) y cuenta ya existente (con enlace a recuperación). Ver 8.7.

### 8.5.4 Stepper

- Cada paso tiene número, nombre y estado.
- Estados [O] (icono + texto, nunca solo color):

| Estado | Icono | Texto accesible | Color |
|---|---|---|---|
| Completado | ✓ | “completado” | `success` |
| Actual | Número resaltado | `aria-current="step"` | `primary` |
| Pendiente | Número | “pendiente” | `neutral-500` |
| Con errores | ! | “con errores” | `danger` |
| No disponible | Candado | “no disponible” | `neutral-500` |

- El porcentaje de avance se muestra en el header. [O] Se calcula como secciones completadas / 10 y se acompaña de texto (“3 de 10 secciones completas”), nunca solo una barra.
- El botón “Guardar progreso” está siempre visible en las secciones editables.
- [O] El estado del stepper proviene del servidor; no se marca una sección como completa por el solo hecho de haberla visitado.

### 8.5.5 Secciones repetibles

- Lista de tarjetas con resumen del ítem (título, institución, fechas y estado del documento adjunto).
- Botones “Editar” y “Eliminar” con nombre accesible específico (“Editar formación: Maestría en…”).
- Botón “Agregar” al final.
- Modal o formulario expandible. [O] En móvil se usa pantalla completa o formulario en página, no un modal flotante.
- Validación al guardar.
- Confirmación al eliminar, nombrando el ítem.
- [O] Estado vacío: “Aún no ha registrado [formación]. Use ‘Agregar’ para incluir la primera.” Si la sección admite responder “No”, la pregunta Sí/No va antes de la lista.
- [O] Si una sección exige un mínimo de ítems **[V]**, se indica antes de la lista.
- [O] Tras guardar o eliminar, el foco vuelve a la tarjeta afectada o al botón “Agregar”, y se anuncia el resultado en una región `aria-live="polite"`.

### 8.5.6 Cargo y elegibilidad (decisión sensible)

- Tres tarjetas de cargo con nombre del cargo y requisitos clave según bases **[V]**; selección con radio buttons dentro de las tarjetas (toda la tarjeta es zona táctil).
- [O] Antes de seleccionar, bloque institucional: “La selección del cargo es definitiva para esta postulación.”
- [O] Confirmación en modal con el nombre del cargo repetido, la consecuencia y botones “Confirmar cargo” / “Revisar otra vez”. El foco inicial va a “Revisar otra vez”.
- **Elegibilidad** se presenta en una **pantalla propia, no en un modal** (debe poder releerse y no cerrarse por error):
  - Resultado en un titular claro: “Cumple los requisitos de elegibilidad” / “No cumple los requisitos de elegibilidad para este cargo”.
  - Motivo concreto (qué requisito no se cumple), sin lenguaje que culpe al ciudadano.
  - Fecha y hora del resultado.
  - [O] Estilo de **resultado administrativo** (bloque institucional), no `danger`.
  - [O] El resultado queda disponible en “Mi postulación”.
  - **[V]** Si existe un mecanismo de revisión o reclamo, su texto y canal los define la institución; la interfaz no lo sugiere si no está aprobado.

### 8.5.7 Inhabilidades (decisión sensible)

- [O] Introducción breve que explique qué es una declaración de inhabilidad y que la respuesta tiene carácter de declaración **[V]** (texto jurídico).
- [O] Las 13 preguntas usan el texto jurídico literal **[V]**, sin reformular. Se permite agregar ayuda en lenguaje sencillo debajo, aprobada por el área jurídica.
- [O] Cada pregunta es un `fieldset` con `legend` y dos radio buttons “Sí” / “No”; **ninguna opción preseleccionada**; no se usan interruptores (toggle).
- [O] Los campos condicionales aparecen inmediatamente debajo de la respuesta que los activa, y se anuncia su aparición.
- [R] Agrupar las preguntas en bloques temáticos de 3–5 con subtítulo, si el área jurídica lo admite **[V]**.

### 8.5.8 Resumen y consentimiento

- Lista de secciones con estado (completa/incompleta/con errores) y enlaces “Corregir” que llevan directamente al campo o ítem.
- [O] Vista de solo lectura de los datos ingresados por sección (desplegable), para revisar antes de enviar, incluido el cargo seleccionado y los documentos adjuntos (nombre, tamaño).
- Cinco consentimientos con checkbox:
  - [O] Ninguno preseleccionado; cada uno con su propia etiqueta y enlace al texto completo.
  - [O] Texto legal a 16 px, legible sin abrir otro documento cuando sea breve.
  - **[V]** Textos y obligatoriedad de cada consentimiento según la normativa de protección de datos personales aplicable y las bases.
- Botón “Enviar postulación”. [O] En lugar de un botón deshabilitado sin explicación, se muestra el botón activo y, al pulsarlo con requisitos pendientes, un resumen de errores que lista qué falta (secciones incompletas, consentimientos sin marcar). Si se mantiene deshabilitado, el texto adyacente indica la causa.
- Confirmación antes de enviar (modal):
  - Título: “¿Enviar su postulación?”
  - Cuerpo: “Después del envío no podrá modificar la información ni los documentos. Cargo: [cargo].”
  - Botones: “Enviar postulación” / “Seguir revisando” (foco inicial en “Seguir revisando”).
- [O] El envío no puede ejecutarse dos veces (botón en estado de carga, petición idempotente en servidor según arquitectura).
- [O] Si el envío falla por error técnico, se informa que **la postulación no fue enviada** y que los datos se conservan; nunca se muestra un estado ambiguo.

### 8.5.9 Enviada (comprobante)

- Código único destacado (`text-2xl`, cifras tabulares, botón “Copiar código”).
- Fecha y hora de envío con zona horaria (“1 de noviembre de 2026, 21:08 (hora de Ecuador)”).
- Estado según el modelo de estados.
- Cargo al que postula.
- Botón “Descargar PDF”.
- Aviso de que el formulario debe imprimirse y firmarse; **[V]** indicar plazo y lugar de entrega solo si las bases lo establecen.
- Enlace a “Mi postulación”.
- [O] La información del comprobante coincide exactamente con la del PDF y con “Mi postulación”.
- **[V]** Mecanismo de verificación de autenticidad del comprobante (p. ej., código verificable): pendiente de definición institucional; la interfaz no debe insinuar una verificación inexistente.
- [R] Si el sistema envía confirmación por correo, indicarlo con la dirección enmascarada.

### 8.5.10 Mi postulación

- [O] Código, cargo, fecha y hora de envío, estado actual con texto explicativo de qué significa y, si aplica, próximo paso.
- [O] Historial de estados con fecha y hora, en orden cronológico inverso.
- [O] Los estados y sus textos provienen del modelo de estados del SDD; la interfaz no crea estados propios.
- Descarga del PDF y del comprobante.
- [O] Si la postulación no se ha enviado, se muestra “Borrador — no enviada” con fecha del último guardado y acceso para continuar.

---

## 8.6 Patrones de formulario

| Patrón | Uso | Requisito |
|---|---|---|
| Campo con label arriba | Todos los formularios. | [O] `label` visible y asociado; nunca solo `placeholder`. |
| Ayuda contextual | Campos legales o técnicos. | [O] Debajo de la etiqueta, vinculada con `aria-describedby`; visible sin interacción cuando sea esencial. |
| Error inline | Validación por campo. | [O] Al salir del campo o al guardar, no mientras se escribe. |
| Resumen de errores | Al intentar guardar. | [O] Al inicio del formulario, con enlaces a cada campo; recibe el foco. |
| Campo read-only | Datos verificados. | [O] Ver 8.6.1. |
| Condicional | Pareja, familiares, remoción, respaldo. | [O] Aparece debajo de la pregunta que lo activa; si se oculta, sus datos no se envían y se informa al usuario. |
| Modal | Agregar/editar ítems repetibles. | [O] Ver 8.8 (foco). Pantalla completa en móvil. |
| Confirmación | Eliminar, enviar, cambiar cargo (si la regla lo permite). | [O] Ver 8.6.4. |
| Autosave | Opcional; no reemplaza “Guardar progreso”. | Ver 8.6.2. |
| Redundancia | Datos ya proporcionados. | [O] No se piden de nuevo; se precargan (WCAG 3.3.7). |

### 8.6.1 Tipos de campo [O]

| Tipo | Señal visual | Señal accesible |
|---|---|---|
| Obligatorio | Texto “(obligatorio)” junto a la etiqueta. | `required` / `aria-required="true"`. |
| Opcional | Por defecto sin marca, salvo que el formulario sea mayoritariamente obligatorio; entonces “(opcional)”. Elegir un criterio único para todo el sistema. | — |
| Verificado | Fondo `neutral-100`, icono de verificado y texto “Dato verificado” + fuente si está aprobada **[V]**. | `readonly`, descripción con la fuente. |
| Solo lectura (no verificado) | Fondo `neutral-100`, sin icono de verificado. | `readonly`. |
| Calculado (p. ej., tiempo de experiencia) | Texto de valor con etiqueta “Calculado automáticamente”. | Región `aria-live="polite"` al actualizarse. |

- [O] Si un dato verificado es incorrecto, se indica a quién recurrir **[V]** (canal de soporte), sin permitir su edición.
- [O] Un asterisco solo, sin texto, no es suficiente para marcar obligatoriedad.

### 8.6.2 Guardado, recuperación y conectividad [O]

- Botones: “Guardar progreso” (secundario: guarda y permanece) y “Guardar y continuar” (primario: guarda y avanza). Se reemplaza el ambiguo “Continuar”.
- Indicador de guardado persistente junto a los botones, con `role="status"`:

| Estado | Texto |
|---|---|
| Sin cambios | “Último guardado: hoy, 14:32.” |
| Cambios pendientes | “Hay cambios sin guardar.” |
| Guardando | “Guardando…” |
| Guardado | “Progreso guardado a las 14:35.” |
| Error | “No se guardaron los cambios. Sus datos siguen en pantalla. Reintente.” |

- El toast “Progreso guardado.” es complementario; el indicador persistente es la fuente de verdad, porque los toasts desaparecen.
- Un fallo de guardado **nunca** borra lo escrito en pantalla.
- Si se implementa autosave, se indica en el mismo indicador y no salta el foco.
- [R] Conservar en el navegador una copia local temporal del formulario en edición para recuperarla tras un corte; al recuperarla, preguntar “Encontramos cambios sin guardar del [fecha y hora]. ¿Desea recuperarlos?”. **[V]** Requiere evaluación de arquitectura y de protección de datos (datos personales en el dispositivo).
- Sin conexión detectada: alerta persistente “Sin conexión. Sus cambios no se pueden guardar hasta que se restablezca.” y botones de guardar en estado informativo.
- [R] Peso de cada pantalla de formulario (HTML+CSS+JS, comprimido) ≤ 300 KB; uso de skeletons solo si la carga supera 1 s.

### 8.6.3 Documentos adjuntos [O]

Antes del selector: tipo y tamaño permitidos (“Solo PDF. Tamaño máximo: 2 MB.” según reglas de negocio) y qué debe contener el documento.

| Estado | Presentación | Acciones |
|---|---|---|
| Sin archivo | Botón “Seleccionar archivo PDF” (input nativo accesible, no solo arrastrar y soltar). | Seleccionar. |
| Subiendo | Barra de progreso con porcentaje en texto, nombre del archivo. | Cancelar. |
| Cargado | Icono, nombre del archivo, tamaño, fecha y hora de carga. | Ver, Reemplazar, Eliminar. |
| Error de validación | Mensaje específico: “El archivo pesa 3,4 MB. El máximo es 2 MB.” / “El archivo no es PDF.” | Seleccionar otro. |
| Error técnico | “No se pudo subir el archivo por un problema de conexión. Reintente.” | Reintentar. |

- La validación de tipo y tamaño ocurre antes de subir, cuando sea posible.
- Arrastrar y soltar, si existe, es una alternativa y no la única vía (WCAG 2.5.7).
- Tras el envío definitivo, los documentos se muestran en solo lectura.
- [R] Indicar cómo reducir el tamaño de un PDF mediante un enlace de ayuda, si soporte lo provee.

### 8.6.4 Confirmaciones de operaciones irreversibles [O]

- Título en forma de pregunta con la acción concreta (“¿Eliminar esta experiencia laboral?”).
- Cuerpo con el objeto afectado y la consecuencia (“No podrá recuperarla.”).
- Botones con verbo: acción (“Eliminar”) y alternativa (“Cancelar”); nunca “Sí/No” ni “Aceptar”.
- El foco inicial va a la opción no destructiva.
- `Esc` equivale a cancelar.

### 8.6.5 Entradas específicas [O]

| Dato | Implementación |
|---|---|
| Cédula | `inputmode="numeric"`, `maxlength="10"`, sin `type="number"`. |
| Fechas | Día/mes/año separados o texto `dd/mm/aaaa` con formato visible. |
| Teléfono, correo | `type="tel"` / `type="email"` y `autocomplete` correspondiente. |
| Dirección | `autocomplete` (`address-line1`, `address-level1`, etc.). |
| País, provincia, cantón | Selects dependientes; al cambiar el padre, el hijo se reinicia y se anuncia. |
| Textos libres | Contador de caracteres solo si existe un límite. |

---

## 8.7 Feedback

### Tipología de mensajes [O]

| Tipo | Origen | Tratamiento | Tono |
|---|---|---|---|
| **Validación** | Dato no cumple formato o regla. | Error inline + resumen, `danger`. | Qué falta y cómo corregirlo. |
| **Error técnico** | Red, servidor, servicio externo. | Alert `danger` o `warning`. | Indicar que no es culpa del usuario, si sus datos se conservaron y qué hacer. |
| **Resultado administrativo** | Aplicación de una regla del concurso (nacionalidad, elegibilidad). | Bloque institucional (`info`), pantalla propia. | Neutral, con fundamento. **No usa `danger`.** |
| **Confirmación** | Acción completada. | Toast `success` + indicador persistente. | Breve. |

### Mensajes

| Situación | Componente | Mensaje |
|---|---|---|
| Guardado | Toast success + indicador | “Progreso guardado.” |
| Error de campo | Texto danger | Mensaje específico: qué falta + cómo corregir (8.10). |
| Error de red | Alert | “No se pudo conectar con el sistema. Sus datos siguen en pantalla. Reintente en unos minutos.” |
| Error del servidor | Alert | “Ocurrió un problema en el sistema. Su información no se perdió. Si persiste, contacte a soporte.” + código de referencia. |
| Servicio de verificación no disponible | Alert warning | “No fue posible verificar su identidad en este momento. Inténtelo más tarde.” |
| Bloqueo | Alert warning | “Por seguridad, el registro está bloqueado temporalmente por varios intentos fallidos. Podrá intentarlo nuevamente a las [hh:mm].” + cuenta regresiva. Duración según seguridad **[V]**. |
| Nacionalidad | Pantalla de resultado institucional | “Este concurso requiere nacionalidad ecuatoriana, según las bases del concurso.” Texto final **[V]**. |
| No elegible | Pantalla de resultado institucional | Resultado + motivo + fecha (8.5.6). Texto **[V]**. |
| Envío | Modal confirm | “¿Enviar su postulación? Después del envío no podrá modificar la información ni los documentos.” |
| Envío exitoso | Pantalla | Comprobante (8.5.9). |
| Sesión por expirar | Modal | “Su sesión terminará en 2 minutos por inactividad. ¿Desea continuar?” → “Continuar sesión”. Tiempo de aviso **[V]** seguridad. |
| Sesión expirada | Redirect al login | “Su sesión terminó por inactividad. Los datos guardados se conservan; ingrese de nuevo para continuar.” |
| Fase cerrada | Bloque institucional | “El periodo de postulación cerró el [fecha]. Ya no es posible enviar postulaciones.” |

Reglas [O]:

- Toasts con `role="status"`, duración mínima 6 s, pausables al pasar el cursor o enfocar, cerrables, y nunca con información que solo exista en el toast.
- Errores con `role="alert"` solo cuando aparecen tras una acción del usuario; no anunciar errores al cargar la página.
- Ningún mensaje culpa al ciudadano (“Usted ingresó mal…”); se describe el dato y la corrección.
- La sesión no expira sin aviso previo con opción de extenderla (WCAG 2.2.1), y la expiración nunca descarta datos ya guardados.
- Los códigos de error técnico se muestran como referencia para soporte, no como mensaje principal.

---

## 8.8 Accesibilidad

Referencia: **WCAG 2.2 AA**. Todos los puntos son [O].

### Perceptible

- Contraste de texto ≥ 4.5:1 (≥ 3:1 en texto grande); componentes, bordes de campo e indicadores ≥ 3:1 (8.2).
- No usar color como único indicador.
- Textos alternativos en iconos informativos; iconos decorativos con `aria-hidden="true"`.
- Estructura semántica: un solo `h1` por pantalla, jerarquía de encabezados sin saltos, listas como `ul/ol`, tablas de datos con `th` y `scope`.
- `lang="es"` en el documento.
- Reflujo sin desplazamiento horizontal a 320 px de ancho CSS (equivale a 1280 px con zoom 400 %).
- Texto ampliable al 200 % sin pérdida de contenido ni funcionalidad.

### Operable

- Navegación por teclado completa, en orden lógico de lectura; sin trampas de foco.
- Foco visible con `ring` (8.2) y no oculto por header, barras fijas ni banners (WCAG 2.4.11).
- Enlace “Saltar al contenido”.
- Títulos de página únicos y descriptivos por paso.
- Áreas táctiles ≥ 44 × 44 px (supera el mínimo de 24 px de WCAG 2.5.8).
- Sin límites de tiempo sin aviso y extensión; la cuenta regresiva de bloqueo no requiere acción del usuario.
- Modales: foco al abrir en el título o en el primer control seguro, foco atrapado dentro, `Esc` para cerrar (salvo que el cierre implique pérdida de datos, en cuyo caso confirma), foco devuelto al control que lo abrió; `role="dialog"`, `aria-modal="true"`, `aria-labelledby`.

### Comprensible

- `label` asociado a cada input; `fieldset`/`legend` para grupos de radio y checkbox.
- `aria-invalid` y `aria-describedby` (ayuda + error).
- `role="alert"` para errores tras una acción.
- Formularios con `autocomplete` adecuado (8.6.5).
- Evitar `placeholder` como única etiqueta.
- Ningún cambio de contexto al enfocar o al cambiar un valor sin aviso (p. ej., no navegar al seleccionar un radio).
- Autenticación accesible: permitir pegar y gestores de contraseñas; ninguna prueba cognitiva adicional (WCAG 3.3.8). **[V]** Si se incorpora un CAPTCHA, debe tener alternativa accesible.
- Lenguaje sencillo (8.1.2).

### Robusto

- Nombre, rol y valor correctos en todos los componentes personalizados; preferir elementos HTML nativos.
- Mensajes de estado anunciados mediante regiones `aria-live` sin mover el foco.

### Matriz mínima de pruebas

| Herramienta | Combinación |
|---|---|
| Análisis automático | axe-core (o equivalente) sin errores críticos ni graves. |
| Lector de pantalla escritorio | NVDA + Chrome o Firefox (Windows). |
| Lector de pantalla móvil | TalkBack + Chrome (Android); VoiceOver + Safari (iOS). |
| Teclado | Recorrido completo de registro → envío sin ratón. |
| Zoom | 200 % y 400 % en escritorio. |

[R] Prueba de usabilidad con al menos 5 personas, incluidas personas con baja alfabetización digital y usuarios de tecnologías de apoyo, antes de la apertura del concurso.

---

## 8.9 Responsive

| Breakpoint | Comportamiento |
|---|---|
| `< 640 px` | Una columna; stepper compacto desplegable; botones full width apilados (primario arriba); modales a pantalla completa; tablas como tarjetas. |
| `640–1024 px` | Dos columnas solo en campos cortos relacionados (p. ej., provincia/cantón); campos largos, preguntas de inhabilidades y adjuntos a una columna; stepper horizontal. |
| `> 1024 px` | Stepper como barra lateral de progreso; contenido a la derecha con ancho `max-w-5xl` (ver 8.3). |

### Reglas

- No scroll horizontal (verificado a 320 px).
- Tablas se convierten en tarjetas en móvil.
- Modales ocupan el ancho disponible; en `< 640 px`, pantalla completa con título y botón “Cerrar” visibles.
- Botones táctiles ≥ 44 px, con separación ≥ 8 px entre controles adyacentes.
- [O] No se comprimen componentes complejos: las tarjetas de cargo se apilan, no se reducen a tres columnas estrechas.
- [O] En móvil, el teclado virtual no debe ocultar el campo enfocado ni los mensajes de error.
- [O] Orientación vertical y horizontal soportadas (WCAG 1.3.4).
- [R] Probar en un dispositivo Android de gama baja y en red 3G simulada.

---

## 8.10 Microcopy

| Contexto | Texto |
|---|---|
| Cédula | “Ingrese su cédula de identidad (10 dígitos).” |
| Fecha de expedición | “Ingrese la fecha de expedición que figura en su cédula (día, mes y año).” |
| Código dactilar | “Ingrese el código dactilar que figura en su cédula.” **[V]** formato (alfanumérico) y ubicación. |
| Contraseña | “Mínimo 10 caracteres, con mayúscula, minúscula, número y símbolo.” (según política de seguridad) |
| Guardado | “Puede guardar su avance y continuar después.” |
| Nacionalidad | “Este concurso requiere nacionalidad ecuatoriana.” |
| Cargo | “La selección del cargo es definitiva para esta postulación.” |
| Documento | “Solo PDF. Tamaño máximo: 2 MB.” |
| Envío | “Revise su información. Después del envío no podrá modificarla.” |
| PDF | “Descargue, imprima y firme su formulario.” |
| Recuperación | “¿Olvidó su contraseña?” |
| Campo obligatorio vacío | “Ingrese [nombre del dato].” |
| Formato incorrecto | “La cédula debe tener 10 números. Revise que no tenga letras ni espacios.” |
| Fecha inválida | “La fecha de fin debe ser posterior a la fecha de inicio.” |
| Lista vacía | “Aún no ha registrado [elemento]. Use ‘Agregar’ para incluir el primero.” |
| Dato verificado | “Dato verificado. No se puede modificar.” |

Reglas [O]: tratamiento de usted; mensajes de error con la estructura **qué ocurre + cómo corregirlo**; mismas palabras para los mismos conceptos (p. ej., siempre “postulación”, nunca alternar con “solicitud” o “aplicación”).

---

## 8.11 Criterios de aceptación visual

Cada pantalla se acepta si cumple todos los criterios [O] aplicables.

**Generales**

- La pantalla carga sin errores de consola.
- El formulario es usable a 360 px y no presenta scroll horizontal a 320 px.
- No hay texto cortado ni solapado con zoom de texto al 200 %.
- Todos los colores provienen de tokens y cumplen los contrastes de 8.2 (verificado con herramienta).
- Hay un solo botón primario por pantalla o modal.
- Ningún estado se comunica solo con color.

**Formularios**

- Los errores son visibles, específicos, anunciados por lector de pantalla y enlazados desde el resumen de errores.
- Cada campo tiene etiqueta visible; obligatorios marcados con texto.
- Los campos verificados y de solo lectura se distinguen visual y programáticamente.
- El botón de guardar muestra estado de carga y no permite doble envío.
- El indicador de guardado muestra la hora del último guardado real.
- Un fallo de red al guardar no borra los datos en pantalla.
- Salir con cambios sin guardar muestra confirmación.
- Los adjuntos muestran los cinco estados de 8.6.3.

**Navegación y estado**

- El stepper refleja el progreso real según el servidor.
- El `<title>` y el foco se actualizan al cambiar de paso.
- El recorrido completo (registro → envío) es posible solo con teclado.
- La sesión avisa antes de expirar y permite extenderla.

**Decisiones sensibles**

- El resultado de elegibilidad aparece en pantalla propia, con motivo y fecha, sin estilo de error, y queda visible en “Mi postulación”.
- La confirmación de cargo y de envío nombra el objeto y la consecuencia, con foco inicial en la opción segura.
- Ningún consentimiento ni respuesta de inhabilidad aparece preseleccionado.

**Comprobante**

- El código, fecha y hora, cargo y estado coinciden en pantalla “Enviada”, PDF y “Mi postulación”.
- El PDF descargado conserva el formato.

**Accesibilidad**

- axe-core sin incidencias críticas ni graves.
- Prueba con NVDA y TalkBack superada en registro, una sección repetible, adjuntos, resumen y envío.

---

## 8.12 Pendientes de validación

| # | Tema | Validación | Sección |
|---|---|---|---|
| 1 | Formato y ubicación del código dactilar (alfanumérico vs “10 dígitos”). | Funcional | 8.5.3, 8.10 |
| 2 | Texto del aviso de tratamiento de datos y de los cinco consentimientos; normativa aplicable. | Jurídica | 8.5.3, 8.5.8 |
| 3 | Texto literal y posible agrupación de las 13 preguntas de inhabilidades; carácter de declaración. | Jurídica | 8.5.7 |
| 4 | Textos de no elegibilidad y nacionalidad; existencia de mecanismo de revisión o reclamo. | Jurídica / institucional | 8.5.6, 8.7 |
| 5 | Etiqueta “Enviar postulación” en lugar de “Guardar postulación”. | Funcional | 8.4, 8.5.8 |
| 6 | Si el cargo puede cambiarse antes del envío (8.6 original listaba “cambiar cargo” pese a ser “definitivo”). | Reglas de negocio | 8.3, 8.6 |
| 7 | Duración de bloqueo, intentos y tiempo de sesión/aviso. | Seguridad | 8.5.2, 8.7 |
| 8 | Mecanismo de verificación de autenticidad del comprobante. | Institucional | 8.5.9 |
| 9 | Plazo y lugar de entrega del formulario impreso y firmado. | Bases del concurso | 8.5.9 |
| 10 | Requisitos visibles en landing y tarjetas de cargo; mínimos de ítems por sección. | Bases del concurso | 8.5.1, 8.5.5, 8.5.6 |
| 11 | Copia local del borrador en el navegador. | Arquitectura / protección de datos | 8.6.2 |
| 12 | Fuente de los datos verificados que se muestra al ciudadano y canal para reportar errores. | Institucional | 8.6.1 |
| 13 | Orden de los 10 pasos y ubicación de Elegibilidad fuera del stepper. | Funcional | 8.3 |
