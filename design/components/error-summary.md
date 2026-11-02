# Resumen de Errores

Componente accesible que lista todos los errores de validación tras un intento de guardado o avance de sección.

---

## Ubicación y Foco

- Aparece en la parte superior del formulario, inmediatamente debajo de la barra sticky de sección.
- Al renderizarse tras un fallo, **recibe el foco programático automáticamente** para anunciar los problemas al lector de pantalla (`tabIndex={-1}`).

---

## Estructura y Estilos

- **Contenedor:**  
  `bg-danger-50 border border-danger border-l-4 rounded-md p-4`, con atributos `role="alert"` y `tabIndex={-1}`.
- **Encabezado:**  
  Título visible en negrita: **"Revise los siguientes datos:"** (`font-semibold text-neutral-900`).
- **Lista de errores:**  
  Lista desordenada (`<ul>`) donde cada elemento es un enlace (`<a>`) que apunta al campo correspondiente (`href="#campo-id"`).
- **Consistencia de texto:**  
  El texto del enlace en el resumen debe coincidir exactamente con el mensaje de error que aparece junto al campo.
