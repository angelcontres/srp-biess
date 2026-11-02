---
name: "Sistema de Postulación BIESS — Concurso Público de Méritos y Oposición"
colors:
  primary: "#0B5CAD"        # acciones principales, header, enlaces
  secondary: "#F2B705"      # acento institucional (solo la franja bajo el header)
  accent: "#083F78"         # primary-dark: hover, énfasis, anillo de foco
  background: "#F8FAFC"     # fondo de página (neutral-50)
  surface: "#FFFFFF"        # tarjetas, campos, modales, barra de sección
  text: "#0F172A"           # texto principal (neutral-900)
  muted: "#475569"          # texto secundario y ayudas (neutral-600, 7.58:1)
  success: "#1E7B45"
  warning: "#B45309"
  danger: "#B42318"
  # Extensiones obligatorias
  info: "#0B5CAD"           # mismo valor que primary; para avisos institucionales
  neutral-100: "#F1F5F9"    # fondo de campos de solo lectura y badges neutros
  neutral-700: "#334155"    # instrucciones
  line: "#E2E8F0"           # divisores INTERNOS (1.23:1, nunca límite de contenedor)
  line-strong: "#7F8DA1"    # límite de contenedores (3.37:1 sobre surface, 3.22:1 sobre background)
  line-input: "#64748B"     # borde de inputs (4.76:1)
  info-50: "#EFF6FF"
  success-50: "#ECFDF3"
  warning-50: "#FFFBEB"
  danger-50: "#FEF3F2"
typography:
  fontFamily:
    heading: "Inter, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    body: "Inter, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    mono: "ui-monospace, SFMono-Regular, Menlo, monospace"  # reservada; no se usa en la interfaz
  scale:
    display: { size: "30px", weight: 700, lineHeight: 1.2 }   # solo título de la landing
    h1:      { size: "24px", weight: 700, lineHeight: 1.25 }  # 20px en < 640px
    h2:      { size: "18px", weight: 600, lineHeight: 1.35 }
    h3:      { size: "16px", weight: 600, lineHeight: 1.4 }
    body:    { size: "16px", weight: 400, lineHeight: 1.5 }
    label:   { size: "16px", weight: 500, lineHeight: 1.5 }
    caption: { size: "14px", weight: 400, lineHeight: 1.5 }   # mínimo absoluto del sistema
spacing:
  base: 4
  scale: [4, 8, 12, 16, 24, 32, 48, 64]
rounded:
  sm: "4px"      # badges            → Tailwind 3.4: `rounded`
  md: "6px"      # botones, inputs   → `rounded-md`
  lg: "8px"      # tarjetas, modales → `rounded-lg`
  full: "9999px" # solo radio buttons, spinner y círculos numerados del stepper
elevation:
  shadow-sm: "0 1px 2px rgba(15,23,42,0.08)"   # tooltips
  shadow-md: "0 4px 8px rgba(15,23,42,0.12)"   # menús desplegables y selects personalizados
  shadow-lg: "0 12px 24px rgba(15,23,42,0.18)" # modales
layout:
  maxWidth: "1024px"          # formularios (max-w-5xl)
  landingMaxWidth: "1280px"   # landing (max-w-7xl)
  sectionGap: "48px"          # entre bloques de contenido
  fieldGap: "24px"            # entre campos
  cardPadding: "24px"         # 16px en < 640px
  controlHeight: "44px"       # inputs, botones y áreas táctiles
  gutterMobile: "16px"
---

# Sistema de Diseño — SRP-BIESS

Este archivo es el índice y punto de entrada al sistema de diseño visual de la postulación al concurso del BIESS. **Léelo antes de escribir cualquier componente.** Cuando una regla de este sistema contradice una costumbre o un valor por defecto (de Tailwind, de librerías externas o de "lo que suele verse bien"), **gana este sistema**.

Las reglas de negocio (plazos, intentos, elegibilidad, estados de la postulación, textos jurídicos) **no están aquí**. Viven en el SDD (`docs/SDD/08-diseno-interfaz-usuario.md` y documentos asociados). No inventes contenido institucional; cuando falte, usa un placeholder visible: `[PENDIENTE: …]`.

---

## Mapa de Lectura por Tarea

Para maximizar la eficiencia y no saturar contexto, lee únicamente los módulos pertinentes a tu tarea:

| Si tu tarea es... | Consulta principalmente: |
|---|---|
| **Crear una nueva pantalla o página global** | [`design/layout.md`](./design/layout.md), [`design/tokens.md`](./design/tokens.md), [`design/components/header-footer.md`](./design/components/header-footer.md) |
| **Construir una sección del formulario de postulación** | [`design/components/section-bar.md`](./design/components/section-bar.md), [`design/components/stepper.md`](./design/components/stepper.md), [`design/components/field.md`](./design/components/field.md), [`design/components/save-status.md`](./design/components/save-status.md), [`design/states.md`](./design/states.md) |
| **Crear o modificar un componente UI base** | [`design/tokens.md`](./design/tokens.md), [`design/components/`](./design/components/), [`design/accessibility.md`](./design/accessibility.md) |
| **Gestionar estados de red, guardado o errores** | [`design/states.md`](./design/states.md), [`design/components/error-summary.md`](./design/components/error-summary.md), [`design/components/alert.md`](./design/components/alert.md) |
| **Redactar microcopy, etiquetas o mensajes** | [`design/content.md`](./design/content.md) |
| **Auditar código de interfaz o validar un PR** | [`design/prohibited.md`](./design/prohibited.md), [`design/accessibility.md`](./design/accessibility.md) |

---

## 0. Asunciones y Decisiones

Decisiones tomadas en entrevista con el responsable del producto:

| # | Decisión |
|---|---|
| 1 | Solo tema claro. Debe funcionar en modo de alto contraste del sistema (`forced-colors`). |
| 2 | Radios: botones e inputs 6 px, tarjetas y modales 8 px, badges 4 px. Clases estándar de Tailwind 3.4.17, sin radios personalizados. |
| 3 | La elevación se expresa con **bordes**; la sombra solo para lo que flota (tooltip, menú, modal). |
| 4 | Densidad intermedia: controles de 44 px, 24 px entre campos, 32–48 px entre bloques. |
| 5 | Referencia de carácter: **GOV.UK**. Prohibido: emojis, el punto medio (·) como separador, mayúsculas sostenidas y efectos de desenfoque o vidrio. |
| 6 | Componente principal: **la barra de sección sticky con línea de progreso sólida**. |
| 7 | Movimiento mínimo; spinner solo dentro de botones; skeleton para cargas de layout. |
| 8 | Errores: backend (500/503) muestra alerta y permite reintentar sin recargar; frontend usa ErrorBoundary + borrador en `sessionStorage`. |
| 9 | Norma WCAG 2.2 AA; objetivos AAA: texto de cuerpo 7:1 y áreas táctiles de 44 px. Regla propia: límites de contenedor ≥ 3:1. |
| 10 | Stack: Tailwind CSS 3.4.17, React, `lucide-react` ^0.474.0. |

Asunciones del redactor (corregir si no aplican):

- `accent` del front matter = `primary-dark` (`#083F78`). La paleta institucional no tiene un tercer color de acento, y `secondary` es el amarillo.
- El estado vacío es solo texto y botón, sin icono ni ilustración.
- El término de navegación es siempre **"Sección"**, nunca "Paso".
- El logo del BIESS es un asset oficial que entrega la institución; hasta tenerlo, usa un placeholder rectangular con el texto "Logo BIESS". No dibujes un logo inventado.

---

## 1. Carácter Visual

Un servicio público, no un producto de consumo. El ciudadano está realizando un trámite de alta trascendencia; la interfaz debe ser **legible, previsible y sobria**, como un formulario oficial rigurosamente estructurado:

- **Jerarquía:** Determinada por tipografía y espacio geométrico, no por color ni ornamentación.
- **Color funcional:** Azul para acción, rojo para error de validación o técnico, ámbar para advertencia, verde para confirmación. El amarillo `secondary` aparece **una sola vez**: en la franja de 4 px bajo el header.
- **Superficies planas:** Paneles blancos sobre un fondo casi blanco (`neutral-50`), delimitados obligatoriamente por bordes visibles (`border-line-strong`).
- **Enlaces claros:** Siempre con subrayado visible.
- **Un botón primario:** Máximo una sola acción principal por pantalla o modal.

*Si una pantalla puede confundirse con una landing de SaaS, un dashboard analítico o una red social, está mal diseñada.*

---

## Resumen de Reglas Estrictas (Lo que NUNCA debe aparecer)

Un resumen rápido de prohibiciones críticas (especificación exhaustiva en [`design/prohibited.md`](./design/prohibited.md)):

1. **Emojis:** Prohibidos en cualquier parte; usa iconos de Lucide con texto explicativo.
2. **Punto medio (`·`):** Prohibido como separador; usa comas, saltos de línea o columnas.
3. **Mayúsculas sostenidas (`uppercase`):** Prohibidas en títulos y etiquetas; usa mayúscula inicial estándar.
4. **Desenfoque (`backdrop-blur` / vidrio):** Prohibido; usa fondos sólidos (`bg-neutral-900/60`).
5. **Degradados y shimmer:** Prohibidos; usa colores de tokens sólidos y transiciones suaves de opacidad (`animate-pulse`).
6. **Sombras para delimitar tarjetas:** Prohibidas; todo contenedor se delimita con `border border-line-strong`.
7. **Radios exagerados (`rounded-xl` o botones píldora):** Prohibidos; usa solo `rounded-md` y `rounded-lg`.
8. **Anillos de foco con `ring`:** Prohibidos; usa siempre `outline` para compatibilidad con alto contraste.
9. **Spinners a pantalla completa:** Prohibidos; usa skeletons para carga de layout y spinners integrados solo dentro de botones.
10. **Imágenes de stock y banners promocionales:** Prohibidos; sobriedad y texto oficial.
11. **Animaciones complejas por scroll o movimiento de tarjetas:** Prohibidas; movimiento mínimo con `motion-safe:`.
12. **Dashboards con tarjetas de métricas inventadas:** Prohibidos; listas limpias y datos reales.
13. **Placeholder como etiqueta:** Prohibido; `<label>` siempre visible.
14. **Toggles para preguntas de Sí/No:** Prohibidos; usa radio buttons explícitos.
15. **Opciones preseleccionadas por defecto en declaraciones:** Prohibidas; el ciudadano debe marcar conscientemente.
16. **Múltiples botones primarios:** Prohibidos; solo uno por vista.
17. **Color rojo (`danger`) para "No elegible":** Prohibido; es un estado institucional informativo (`info`).
18. **Truncar con `…`:** Prohibido en botones o títulos de sección; se debe permitir salto de línea.
19. **Fuentes monoespaciadas para códigos:** Prohibidas; usa Inter con clase `tabular-nums`.
20. **Flechas `→` en botones:** Prohibidas; redacta verbo + objeto claro.
21. **Paletas de color por defecto de Tailwind:** Prohibidas; usa únicamente los tokens definidos en este kit.
