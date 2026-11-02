# Tarjetas

Contenedores de superficie para agrupar información relevante, formularios de acceso e ítems repetibles.

---

## Variantes de Tarjetas

### 1. Tarjeta de Login y Registro
- Centrada en pantalla sobre fondo general (`bg-background`).
- Clases: `max-w-md w-full bg-surface border border-line-strong rounded-lg p-6 sm:p-8`.
- **Sin sombras**: el límite se delimita estrictamente mediante el borde visible `border-line-strong`.

### 2. Ítem Repetible (Formación, Experiencia, Publicaciones, Capacitación)
- Contenedor de bloque: `bg-surface border border-line-strong rounded-lg p-4 sm:p-6`.
- Título principal con encabezado `h3` (`text-base font-semibold text-neutral-900`).
- Metadatos complementarios en tipografía caption (`text-sm text-muted`).
- Estado de verificación o documento adjunto comunicado mediante un badge accesible.
- Acciones de "Editar" y "Eliminar" en estilo botón terciario, con nombre accesible descriptivo (ej. *"Editar formación: Maestría en Finanzas"*).

### 3. Cargo Seleccionable
- Toda la superficie de la tarjeta actúa como la etiqueta interactiva (`<label>`) de un botón de radio.
- **Estado No Seleccionado:** `border-2 border-line-strong bg-surface`.
- **Estado Seleccionado:** `border-2 border-primary bg-info-50` acompañado del texto visible e indicador *"Seleccionado"*.
- **Prevención de salto de layout:** El borde de 2 px (`border-2`) debe estar presente en ambos estados para evitar oscilaciones de tamaño al seleccionar.
- **Responsive:** En pantallas menores a 1024 px (`< 1024px`), las tarjetas de cargo se apilan verticalmente; nunca forzar tres columnas estrechas.
