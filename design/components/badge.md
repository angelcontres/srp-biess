# Badge de Estado

Etiquetas compactas para indicar el estado de avance, verificación o validez de ítems y secciones.

---

## Estructura Base

`inline-flex items-center gap-1 rounded px-2 py-0.5 text-sm font-medium`

---

## Variantes de Estado

| Estado | Clases Tailwind | Icono (size=16) |
|---|---|---|
| **Completa / Enviada** | `bg-success-50 text-success` | `CircleCheck` |
| **Incompleta** | `bg-warning-50 text-warning` | `TriangleAlert` |
| **Con errores** | `bg-danger-50 text-danger` | `CircleAlert` |
| **Verificado** | `bg-info-50 text-primary` | `ShieldCheck` |
| **Sin iniciar / Borrador** | `bg-neutral-100 text-neutral-700` | — |

---

## Reglas

- **Accesibilidad:** Los badges siempre deben combinar un icono y texto explicativo (nunca color en solitario).
- **Radio:** Utilizan el radio mínimo del sistema: `rounded` (4 px).
