# Toast (Notificación Flotante)

Mensaje flotante no bloqueante para confirmar eventos inmediatos (ej. guardado temporal exitoso).

---

## Directrices de Uso

- **Rol complementario:** Actúa como refuerzo del indicador de guardado persistente; **nunca debe contener información crítica o única** que no esté visible en otra parte de la interfaz.
- **Accesibilidad y persistencia:**  
  - Atributo `role="status"`.  
  - Duración mínima en pantalla: **6 segundos**.  
  - Debe pausar su temporizador de desaparición cuando recibe el puntero (`hover`) o el foco de teclado.  
  - Debe incluir un botón de cierre explícito (`X`).  
  - Solo se muestra **un toast a la vez** en pantalla.

---

## Posicionamiento y Estilos

- **Posición responsive:**  
  - En `< 640 px`: abajo al centro de la pantalla.  
  - En `≥ 640 px`: arriba a la derecha, ubicado inmediatamente debajo de la barra sticky de sección.
- **Clases Tailwind:**  
  `bg-surface border border-line-strong border-l-4 border-l-success rounded-md shadow-md p-4 flex items-center justify-between gap-3`
