# Modales y Diálogos

Ventanas modales para captura de información detallada (ej. formularios de ítems) y confirmación de acciones irreversibles.

---

## Estructura y Accesibilidad

- **Semántica:** Atributos obligatorios `role="dialog"`, `aria-modal="true"`, `aria-labelledby="modal-title"`.
- **Overlay:** Fondo oscurecido sólido `bg-neutral-900/60`, **terminantemente sin `backdrop-blur`**.
- **Gestión del foco:** Al abrir, el foco viaja automáticamente al título del modal o a la opción más segura. El foco queda atrapado dentro del contenedor mientras permanezca abierto. Al pulsar la tecla `Esc`, el modal se cierra y el foco retorna al control interactivo que lo abrió.

---

## Comportamiento Responsive

- **Escritorio y Tablet (≥ 640 px):**  
  `max-w-lg w-full bg-surface border border-line-strong rounded-lg shadow-lg p-6`, centrado en pantalla.
- **Móvil (< 640 px):**  
  Pantalla completa fija (`fixed inset-0 rounded-none bg-surface`), con cabecera superior persistente que contiene el título y un botón de cierre (`X`, 44 px de área táctil).
- **Disposición de botones de acción:**  
  - En ≥ 640 px: alineados a la derecha, ubicando la opción secundaria/cancelar a la izquierda y el botón primario a la derecha.  
  - En < 640 px: botones apilados verticalmente a ancho completo (`w-full`), colocando el botón primario arriba.

---

## Contenido Más Alto que la Ventana

- El panel nunca supera la altura de la ventana (`sm:max-h-full` dentro del overlay con `sm:p-4`; `h-full` en móvil).
- Cabecera (título y cierre) y pie (botones) con `shrink-0`: siempre visibles.
- Solo el cuerpo se desplaza: `min-h-0 flex-1 overflow-y-auto`, con su propia barra de scroll.

---

## Diálogos de Confirmación Irreversible

Para acciones críticas (eliminar un ítem, enviar la postulación definitiva):

- **Título en pregunta directa:** Formulado de forma concreta (ej. *"¿Eliminar esta experiencia laboral?"*, *"¿Enviar su postulación definitiva?"*).
- **Cuerpo explicativo:** Detalla las consecuencias exactas del acto de confirmación.
- **Botones con verbos explícitos:** Usar pares claros como "Eliminar" / "Cancelar", o "Enviar postulación" / "Seguir revisando". Nunca "Aceptar" o "Sí/No".
- **Foco inicial:** Por seguridad, el foco inicial debe dirigirse siempre a la **opción no destructiva** (cancelar).
