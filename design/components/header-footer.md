# Header y Footer

Estructuras de navegación global que enmarcan todas las pantallas del sistema de postulación.

---

## 1. Header (Cabecera Global)

- **Fondo y dimensiones:**  
  `bg-primary text-white h-16 (64px) flex items-center justify-between px-4 sm:px-6`. No es sticky (se desplaza con el scroll natural de la página).
- **Elementos requeridos:**  
  - Logo oficial del BIESS (o placeholder textual hasta disponer del asset).  
  - Nombre oficial del concurso: *"Concurso Público de Méritos y Oposición"*.  
  - Identificación del usuario autenticado (nombre o cédula).  
  - Botón de cierre de sesión: *"Salir"* acompañado del icono `LogOut`.
- **Franja de acento institucional:**  
  Inmediatamente debajo del header se ubica una franja continua de 4 px:  
  `<div className="h-1 bg-secondary w-full" />`  
  *(Es el único lugar de toda la interfaz donde se permite utilizar el color `secondary` amarillo).*
- **Salto accesible:**  
  El primer elemento enfocable de la página debe ser un enlace invisible que se vuelve visible al foco:  
  `<a href="#main-content" className="sr-only focus:not-sr-only focus:p-2 focus:bg-surface focus:text-primary">Saltar al contenido</a>`

---

## 2. Footer (Pie de Página)

- **Fondo y bordes:**  
  `bg-surface border-t border-line-strong py-6 px-4 sm:px-6 text-sm text-muted`.
- **Enlaces requeridos:**  
  Enlaces con texto subrayado (`underline underline-offset-4`) a:  
  - Ayuda  
  - Políticas de privacidad  
  - Soporte técnico  
  - Bases legales del concurso  
- **Consistencia:** Mantienen siempre la misma posición y orden en todas las pantallas del sistema.
