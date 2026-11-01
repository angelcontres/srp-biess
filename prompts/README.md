# Prompt Packs de Ahorro Máximo de Tokens (Equipo SRP-BIESS)

> [!TIP]
> **REGLA DE ORO DEL EQUIPO:**  
> **NUNCA** subas el archivo `requerimiento.pdf` a ChatGPT ni a Claude. Ese PDF pesa 400 KB y se come toda tu cuota gratuita en 2 mensajes.  
> En su lugar, copia y pega el texto de la plantilla del ticket que te toca trabajar. Cada una pesa menos de **300 tokens**.

---

## Índice de Plantillas Rápidas para el Equipo:

* **Paulo (Backend Lead):**
  * [`prompts/SRP-001.md`](./SRP-001.md) — Inicialización Go y Clean Arch
  * [`prompts/SRP-002.md`](./SRP-002.md) — Composition Root e Inyección Manual
  * [`prompts/SRP-003.md`](./SRP-003.md) — Autenticación JWT y Middleware
* **Angel (Data Lead):**
  * [`prompts/SRP-004.md`](./SRP-004.md) — Repositorios pgxpool y Transacciones
* **Carlos (Frontend Lead):**
  * [`prompts/SRP-012.md`](./SRP-012.md) — Setup React + Vite + Tailwind
* **Jose Daniel (Cache & QA Lead):**
  * [`prompts/SRP-007.md`](./SRP-007.md) — Repositorio Redis y TTL

---

## Cómo usarlo con cualquier IA:
1. Abre tu IA favorita (ChatGPT, Claude, Cursor, Copilot o Gemini).
2. Abre la tarjeta del ticket correspondiente (ej. `prompts/SRP-001.md`).
3. Copia las 15 líneas y pégaselas a tu IA.
4. Cuando te entregue el código, corre en tu terminal:
   ```cmd
   verify.bat
   ```
5. Si pasa: ¡Listo para hacer commit! Si falla: solo cópiale las 2 líneas de error a tu IA.
