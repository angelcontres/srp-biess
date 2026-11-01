# 08. Diseño de interfaz de usuario — SDD

## 8.1 Principios de diseño

1. **Claridad:** el postulante siempre sabe en qué sección está, qué falta y qué debe hacer.
2. **Progresividad:** formulario largo dividido en secciones con guardado independiente.
3. **Confianza:** mensajes institucionales, datos verificados y confirmaciones antes de acciones irreversibles.
4. **Accesibilidad:** navegación por teclado, contraste, etiquetas y errores anunciados.
5. **Responsive:** experiencia usable en móvil, tablet y escritorio.
6. **Consistencia:** mismos patrones de campos, botones, alertas y loaders.

---

## 8.2 Sistema de diseño

### Paleta

| Token | Valor | Uso |
|---|---|---|
| `primary` | `#0B5CAD` | Acciones principales, header. |
| `primary-dark` | `#083F78` | Hover, énfasis. |
| `secondary` | `#F2B705` | Acentos institucionales. |
| `success` | `#1E7B45` | Confirmaciones. |
| `warning` | `#B45309` | Advertencias. |
| `danger` | `#B42318` | Errores y bloqueos. |
| `neutral-50` | `#F8FAFC` | Fondo. |
| `neutral-900` | `#0F172A` | Texto principal. |

### Tipografía

- Familia: `Inter`, `system-ui`, sans-serif.
- Títulos: 600–700.
- Cuerpo: 400–500.
- Tamaños: `text-sm` (14 px), `text-base` (16 px), `text-lg` (18 px), `text-2xl` (24 px).

### Espaciado

- Escala Tailwind: 2, 4, 6, 8, 12, 16.
- Contenedor máximo: `max-w-5xl` para formularios; `max-w-7xl` para landing.
- Radios: `rounded-lg` (8 px), `rounded-xl` (12 px).
- Sombras suaves para tarjetas.

### Componentes base

- Botón primario, secundario, terciario y peligro.
- Input, select, textarea, date, radio, checkbox.
- Card, alert, badge, modal, tooltip, toast.
- Stepper, tabs, breadcrumb.
- Skeleton y spinner.

---

## 8.3 Layout y navegación

```text
┌──────────────────────────────────────────────┐
│ Header: logo BIESS | concurso | usuario      │
├──────────────────────────────────────────────┤
│ Stepper: 1 2 3 4 5 6 7 8 9 10               │
├──────────────────────────────────────────────┤
│ Contenido de la sección                      │
│                                              │
│ [Guardar progreso]        [Continuar]        │
├──────────────────────────────────────────────┤
│ Footer: ayuda | privacidad | soporte         │
└──────────────────────────────────────────────┘
```

### Navegación

- El stepper muestra secciones completadas, actual y pendientes.
- Se puede volver a una sección anterior sin perder datos.
- No se puede avanzar si la sección actual tiene errores bloqueantes.
- El resumen final muestra el estado de cada sección.
- En móvil, el stepper se convierte en una barra de progreso compacta.

---

## 8.4 Inventario de pantallas

| Pantalla | Elementos clave | Acciones |
|---|---|---|
| **Landing** | Título, descripción, requisitos, botones. | Iniciar sesión, registrarse. |
| **Login** | Cédula, contraseña, recuperación. | Ingresar. |
| **Registro** | Cédula, tipo de validación, valor, correo, contraseña. | Validar, registrar. |
| **Bloqueo** | Mensaje, cuenta regresiva. | Reintentar al expirar. |
| **Recuperación** | Cédula, validación. | Enviar enlace. |
| **Restablecer** | Nueva contraseña, confirmación. | Guardar. |
| **Personal** | Datos verificados, estado civil, pareja. | Guardar progreso, continuar. |
| **Domicilio** | País, provincia, cantón, dirección, contacto. | Guardar, continuar. |
| **Familiares** | Sí/No, lista, modal de agregar. | Agregar, editar, eliminar. |
| **Cargo** | Tres tarjetas de cargo, confirmación. | Seleccionar, confirmar. |
| **Elegibilidad** | Resultado, mensaje institucional. | Continuar o cerrar. |
| **Formación** | Lista, modal, PDF. | Agregar, editar, eliminar. |
| **Experiencia** | Lista, fechas, tiempo calculado, PDF. | Agregar, editar, eliminar. |
| **Publicaciones** | Sí/No, lista, PDF. | Agregar, editar, eliminar. |
| **Capacitación** | Lista, PDF. | Agregar, editar, eliminar. |
| **Inhabilidades** | 13 preguntas, campos condicionales. | Guardar, continuar. |
| **Resumen** | Estado de secciones, consentimientos. | Guardar postulación. |
| **Enviada** | Código, fecha, PDF. | Descargar, volver. |
| **Mi postulación** | Estado, código, PDF. | Descargar. |

---

## 8.5 Especificación de pantallas clave

### 8.5.1 Landing

- Hero con título “Concurso Público de Méritos y Oposición — Directorio del BIESS”.
- Texto breve del proceso y requisitos.
- Botones “Iniciar sesión” y “Registrarse”.
- Aviso si la fase está cerrada.
- Enlaces a bases, privacidad y soporte.

### 8.5.2 Login

- Campo cédula (10 dígitos).
- Campo contraseña con mostrar/ocultar.
- Enlace “¿Olvidaste tu contraseña?”.
- Botón “Ingresar”.
- Mensajes de error genéricos.
- Estado de carga en el botón.

### 8.5.3 Registro

- Paso 1: cédula + tipo de validación (fecha o código dactilar) + valor.
- Paso 2: correo + contraseña + confirmación.
- Indicador de fortaleza de contraseña.
- Mensaje de bloqueo con cuenta regresiva.
- Al éxito, redirige al login.

### 8.5.4 Stepper

- Cada paso tiene número, nombre y estado.
- Estados: `completado` (check), `actual` (resaltado), `pendiente` (gris), `error` (rojo).
- El porcentaje de avance se muestra en el header.
- El botón “Guardar progreso” está siempre visible.

### 8.5.5 Secciones repetibles

- Lista de tarjetas con resumen del ítem.
- Botones “Editar” y “Eliminar”.
- Botón “Agregar” al final.
- Modal o formulario expandible.
- Validación al guardar.
- Confirmación al eliminar.

### 8.5.6 Resumen y consentimiento

- Lista de secciones con estado (completa/incompleta).
- Enlaces para corregir.
- Cinco consentimientos con checkbox.
- Botón “Guardar postulación” deshabilitado hasta aceptar.
- Mensaje de confirmación antes de enviar.

### 8.5.7 Enviada

- Código único destacado.
- Fecha de envío.
- Estado.
- Botón “Descargar PDF”.
- Aviso de que el formulario debe imprimirse y firmarse.
- Enlace a “Mi postulación”.

---

## 8.6 Patrones de formulario

| Patrón | Uso |
|---|---|
| Campo con label arriba | Todos los formularios. |
| Ayuda contextual | Campos legales o técnicos. |
| Error inline | Validación por campo. |
| Resumen de errores | Al intentar guardar. |
| Campo read-only | Datos verificados. |
| Condicional | Pareja, familiares, remoción, respaldo. |
| Modal | Agregar/editar ítems repetibles. |
| Confirmación | Eliminar, enviar, cambiar cargo. |
| Autosave | Opcional; no reemplaza “Guardar progreso”. |

---

## 8.7 Feedback

| Situación | Componente | Mensaje |
|---|---|---|
| Guardado | Toast success | “Progreso guardado.” |
| Error de campo | Texto danger | Mensaje específico. |
| Error de red | Alert | “No se pudo conectar. Reintente.” |
| Bloqueo | Alert warning | “Registro fallido, inténtelo nuevamente en 10 minutos.” |
| Nacionalidad | Modal danger | Mensaje de cierre. |
| No elegible | Modal danger | Mensaje institucional. |
| Envío | Modal confirm | “¿Está seguro de enviar? No podrá modificar.” |
| Envío exitoso | Pantalla | Código y PDF. |
| Sesión expirada | Redirect | “Su sesión expiró.” |

---

## 8.8 Accesibilidad

- Contraste mínimo 4.5:1.
- Foco visible con `ring`.
- `label` asociado a cada input.
- `aria-invalid` y `aria-describedby`.
- `role="alert"` para errores.
- Navegación por teclado completa.
- No usar color como único indicador.
- Textos alternativos en iconos.
- Formularios con `autocomplete` adecuado.
- Evitar `placeholder` como única etiqueta.

---

## 8.9 Responsive

| Breakpoint | Comportamiento |
|---|---|
| `< 640 px` | Una columna; stepper compacto; botones full width. |
| `640–1024 px` | Dos columnas en formularios; stepper horizontal. |
| `> 1024 px` | Layout con sidebar de progreso; contenido centrado. |

### Reglas

- No scroll horizontal.
- Tablas se convierten en tarjetas en móvil.
- Modales ocupan el ancho disponible.
- Botones táctiles ≥ 44 px.

---

## 8.10 Microcopy

| Contexto | Texto |
|---|---|
| Cédula | “Ingrese su cédula de identidad (10 dígitos).” |
| Fecha de expedición | “Ingrese la fecha de expedición de su cédula.” |
| Código dactilar | “Ingrese el código dactilar de 10 dígitos.” |
| Contraseña | “Mínimo 10 caracteres, con mayúscula, minúscula, número y símbolo.” |
| Guardado | “Puede guardar su avance y continuar después.” |
| Nacionalidad | “Este concurso requiere nacionalidad ecuatoriana.” |
| Cargo | “La selección del cargo es definitiva para esta postulación.” |
| Documento | “Solo PDF. Tamaño máximo: 2 MB.” |
| Envío | “Revise su información. Después del envío no podrá modificarla.” |
| PDF | “Descargue, imprima y firme su formulario.” |

---

## 8.11 Criterios de aceptación visual

- La pantalla carga sin errores de consola.
- El formulario es usable a 360 px.
- Los errores son visibles y accesibles.
- El botón de guardar muestra estado de carga.
- El stepper refleja el progreso real.
- El PDF descargado conserva el formato.
- No hay texto cortado ni scroll horizontal.
