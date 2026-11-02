# Botones

Directrices y variantes de botones para el sistema de postulación.

---

## Base Común

Todos los botones comparten la estructura base:  
`inline-flex items-center justify-center gap-2 h-11 px-5 rounded-md text-base font-semibold`

---

## Variantes

| Variante | Clases Tailwind | Uso |
|---|---|---|
| **Primario** | `bg-primary text-white hover:bg-primary-dark` | Una sola vez por pantalla: "Guardar y continuar", "Enviar postulación", "Ingresar". |
| **Secundario** | `bg-surface text-primary border border-primary hover:bg-info-50` | Acciones secundarias: "Guardar progreso", "Agregar", "Seguir revisando". |
| **Terciario** | `text-primary underline underline-offset-4 hover:text-primary-dark h-auto px-0` | Enlaces en forma de botón: "Sección anterior", "Ver documento". |
| **Peligro** | `bg-danger text-white hover:brightness-90` | Solo dentro de la confirmación de operaciones irreversibles (eliminar). |

---

## Reglas de Interacción y Accesibilidad

- **Foco visible:**  
  `focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark`.  
  **Usa `outline`, nunca `ring`:** `ring` utiliza `box-shadow`, el cual desaparece en el modo de alto contraste del sistema (`forced-colors`). Sobre fondo `primary` (header), usar `outline-white`.
- **Estado de carga:**  
  Icono `LoaderCircle` (`motion-safe:animate-spin`) + texto descriptivo ("Guardando…", "Enviando…"), atributo `aria-disabled="true"`. El botón **mantiene su ancho original** y bloquea clics secundarios.
- **Estado deshabilitado:**  
  Evitar deshabilitar botones. Si resulta imprescindible: `bg-neutral-100 text-muted border border-line-input`, acompañado de texto cercano que explique claramente por qué no está disponible.
- **Microcopy:**  
  Texto imperativo y concreto: **verbo + objeto** (ej. "Guardar progreso", "Enviar postulación"). Nunca usar "Aceptar", "OK", "Enviar" a secas, ni caracteres o flechas tipo `→`.
- **Adaptación móvil (< 640 px):**  
  Los botones situados al final del formulario pasan a ocupar el ancho completo (`w-full`). El botón de la barra de sección sticky mantiene su tamaño compacto.
