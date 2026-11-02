# Documento Adjunto (File Upload)

Componente para la selección, validación y carga de documentos de soporte en formato PDF.

---

## Instrucciones Previas

Antes del control, se debe mostrar siempre un texto en tipografía caption (`text-sm text-muted`):  
*"Solo PDF. Tamaño máximo: 2 MB."* (o 4 MB según la sección) detallando qué contenido debe incluir el documento.

---

## Estados de la Carga

| Estado | Presentación Visual | Acciones Disponibles |
|---|---|---|
| **Sin archivo** | Botón secundario *"Seleccionar archivo PDF"* (icono `Upload`), asociado a un input nativo `<input type="file" accept="application/pdf">` accesible por teclado. | Seleccionar |
| **Subiendo** | Nombre del archivo + barra de progreso sólida (`bg-primary` sobre `bg-line`) + texto de porcentaje *"45 %"*. | Cancelar |
| **Cargado** | Contenedor con borde: `border border-line-strong rounded-md p-3 flex items-center justify-between`: icono `FileText`, nombre del archivo, tamaño en KB/MB, y fecha y hora de carga. | Ver, Reemplazar, Eliminar |
| **Error de validación** | Mensaje específico de advertencia en `text-danger` (ej. *"El archivo pesa 3,4 MB. El tamaño máximo permitido es 2 MB."*). | Seleccionar otro |
| **Error técnico** | Alerta de error compacta indicando fallo de red o servidor. | Reintentar (`RotateCw`) |

---

## Reglas

- La funcionalidad de arrastrar y soltar (*drag and drop*) solo se permite como complemento opcional, **nunca como la única vía** de carga.
- Los archivos PDF se suben inmediatamente al servidor tras su selección; no forman parte del borrador local en `sessionStorage` (ver `design/states.md`).
