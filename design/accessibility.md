# Accesibilidad (WCAG 2.2 AA)

Norma obligatoria: **WCAG 2.2 AA**. Objetivos AAA: texto de cuerpo 7:1 y áreas táctiles mínimas de 44 px.

---

## Directrices de Accesibilidad

- **Contraste de color:** Texto ≥ 4.5:1 (cuerpo ≥ 7:1); inputs, iconos informativos, foco y **límites de contenedor** ≥ 3:1.
- **Foco visible:** Siempre con `outline` (ver `design/tokens.md`), nunca oculto por la barra sticky (`scroll-padding-top`).
- **Navegación por teclado:** Todo operable con teclado, en orden natural de lectura; sin trampas de foco fuera de los modales.
- **Landmarks semánticos:** `header`, `nav`, `main`, `footer`. Un único `h1` por pantalla, sin saltos en la jerarquía de encabezados.
- **Idioma y título:** Atributo `lang="es"` en `html`. Etiqueta `<title>` única y descriptiva por pantalla (ej. "Sección 3 de 10: Familiares | Postulación BIESS").
- **Anuncios para lectores de pantalla:** Mensajes de estado y contadores con `aria-live="polite"`; errores tras una acción con `role="alert"`.
- **Alto contraste del sistema (`forced-colors`):** Bordes reales (no sombras) en todos los contenedores e inputs, foco con `outline`, iconos en `currentColor`.
- **Reflujo y zoom:** Sin scroll horizontal a 320 px; texto escalable al 200 % sin pérdida de información ni superposición.
- **Control de sesión:** Aviso previo antes de que expire la sesión con opción explícita "Continuar sesión"; la expiración nunca borra datos ya guardados.
- **Pruebas mínimas antes de aceptar una pantalla:**
  1. Auditoría axe sin incidencias críticas ni graves.
  2. Recorrido interactivo completo operado exclusivamente con teclado.
  3. Verificación con NVDA + Chrome y TalkBack + Chrome.
