# Contenido y Voz

Este archivo define las reglas de tono, redacción, terminología fija y formateo para todas las pantallas del sistema de postulación.

---

## Reglas de Redacción

- **Tratamiento:** Trato de **usted** estricto en toda la interfaz y comunicaciones del sistema.
- **Estructura de oraciones:** Oraciones cortas, voz activa y mayúscula inicial. Prohibidas las mayúsculas sostenidas en cualquier elemento.
- **Consistencia verbal:** Mismo nombre para la misma acción en el botón, el diálogo de confirmación y el mensaje de resultado (ej. "Enviar postulación" → "¿Enviar su postulación?" → "Postulación enviada").
- **Mensajes de error:** Formato constructivo: **qué ocurre + cómo corregirlo**, sin atribuir culpa al usuario (ej. *"La fecha de fin debe ser posterior a la fecha de inicio."*, nunca *"Fecha inválida"*).
- **Términos oficiales fijos:**
  - **"postulación"** (nunca "solicitud" ni "aplicación").
  - **"sección"** (nunca "paso").
  - Acciones exactas: **"Guardar progreso"**, **"Guardar y continuar"**, **"Enviar postulación"**.
- **Fechas y números:** Formateados siempre con la API `Intl` en locale `es-EC` y zona horaria `America/Guayaquil` (ej. *"1 de noviembre de 2026, 21:08"*). Hora en formato de 24 horas. Prohibido formatear fechas a mano.
- **Separadores visuales:** Usar coma, punto, dos puntos o salto de línea. **Nunca el punto medio (`·`)**, ni la estructura *"palabra — fragmento"* como etiqueta.
- **Sobriedad institucional:** Sin emojis, sin signos de exclamación en mensajes del sistema y sin lenguaje promocional o marketinero.
- **Textos legales:** Los consentimientos, cuestionarios de inhabilidades y resultados de elegibilidad se copian de forma literal desde el SDD o se dejan señalizados como `[PENDIENTE: texto jurídico]`.
