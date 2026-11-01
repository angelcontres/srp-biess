# Agent Harness & AI Software Factory (SRP-BIESS)

Sistema autónomo de desarrollo basado en la tríada de arquitectura de arneses:
1. **Orquestador Cíclico:** **LangGraph** (Feedback loops y máquinas de estado).
2. **Capas de Seguridad:** **Guardrails AI & Pydantic** (Validación de reglas arquitectónicas limpias ADR-003 y contención de rutas).
3. **Conectividad:** **MCP / Tool Protocol** (Inspección del backlog `Docs/TICKETS.md`, gestión de archivos y ejecución de terminal para compilación y pruebas).

---

## 1. Instalación Rápida

Crea un entorno virtual de Python e instala las dependencias:

```bash
cd harness
python -m venv .venv
# En Windows:
.venv\Scripts\activate
# En Linux / Mac:
source .venv/bin/activate

pip install -r requirements.txt
```

---

## 2. Configuración del Proveedor de IA

Copia el archivo de ejemplo y coloca tus credenciales:

```bash
cp .env.example .env
```

Edita `.env` y define tu proveedor:
* **Google Gemini:** `AI_PROVIDER=gemini` y `GEMINI_API_KEY=tu_clave`
* **DeepSeek / OpenRouter:** `AI_PROVIDER=deepseek` y `OPENROUTER_API_KEY=tu_clave`
* **Anthropic Claude:** `AI_PROVIDER=anthropic` y `ANTHROPIC_API_KEY=tu_clave`
* **OpenAI:** `AI_PROVIDER=openai` y `OPENAI_API_KEY=tu_clave`
* **Ollama (Local sin costo):** `AI_PROVIDER=ollama` y `AI_MODEL=qwen2.5-coder:14b`

---

## 3. Uso de la Fábrica de Software

### A) Listar todos los tickets disponibles en el backlog:
```bash
python -m harness.main list
```

### B) Ejecutar la resolución autónoma de un ticket:
```bash
python -m harness.main run SRP-001
```

O especificando el modelo al vuelo:
```bash
python -m harness.main run SRP-001 --provider deepseek --model deepseek-chat
python -m harness.main run SRP-001 --provider gemini --model gemini-2.5-flash
```

---

## 4. Ciclo de Vida del Arnés (StateGraph)

```
[PLANNER] -> Lee el ticket y Docs/SDD -> Produce plan técnico estructurado
   ↓
[CODER]   -> Genera operaciones de archivo
   ↓
[GUARDRAIL] -> Verifica que no haya escape de rutas ni dependencias prohibidas
   ↓
[TESTER]  -> Ejecuta 'go build ./...' o 'npm test' en la terminal
   ↓
¿Pasó la prueba?
   ├── SÍ  -> [COMMITTER] -> Crea rama 'feature/...' y hace Git Commit -> FIN
   └── NO  -> [CODER con Feedback de Error] -> Reintenta corrigiendo el fallo
```
