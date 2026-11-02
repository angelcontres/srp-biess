# Alertas y Avisos

Mensajes contextuales persistentes para comunicar estados relevantes, avisos institucionales y fallos técnicos.

---

## Tipos de Alerta

| Tipo | Clases Tailwind | Icono | Uso |
|---|---|---|---|
| **Institucional** | `bg-info-50 border border-primary border-l-4` | `Info` | Reglas del concurso, resultados administrativos, aviso de fase cerrada. |
| **Error** | `bg-danger-50 border border-danger border-l-4` | `CircleAlert` | Fallos técnicos de backend, errores de guardado de sección. |
| **Advertencia** | `bg-warning-50 border border-warning border-l-4` | `TriangleAlert` | Avisos de bloqueo temporal, intentos restantes, cambios sin guardar. |
| **Éxito** | `bg-success-50 border border-success border-l-4` | `CircleCheck` | Confirmaciones persistentes en pantalla. |

---

## Reglas de Implementación

- **Estructura base:**  
  Todas las alertas utilizan `rounded-md p-4 flex gap-3`, con un borde lateral izquierdo reforzado (`border-l-4`).
- **Color y legibilidad:**  
  El color identificador se restringe al borde y al icono. El cuerpo del texto siempre permanece oscuro (`text-neutral-900`) para garantizar el máximo contraste y legibilidad.
- **Jerarquía:**  
  Admiten un título opcional en negrita (`font-semibold`) y un cuerpo descriptivo claro con recomendaciones de acción cuando corresponda.
