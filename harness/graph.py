"""
Grafo de Estados Cíclico de LangGraph (El Núcleo del Arnés).
Implementa el bucle cerrado: Planificar -> Codificar -> Validar Guardrails -> Probar -> Auto-corregir -> Commit.
"""

import json
from pathlib import Path
from typing import Dict, Any, List

from langgraph.graph import StateGraph, END
from langchain_core.messages import SystemMessage, HumanMessage

from harness.models import HarnessState, CodePlan, FileOperation
from harness.guardrails import HarnessGuardrails
from harness.tools import HarnessToolchain
from harness.llm import get_configured_llm


def create_software_factory_graph(root_dir: Path, provider: str = None, model: str = None):
    """Construye y compila el StateGraph de LangGraph."""
    
    guardrails = HarnessGuardrails(root_dir)
    tools = HarnessToolchain(root_dir)
    llm = get_configured_llm(provider, model)

    # -------------------------------------------------------------
    # NODO 1: PLANIFICADOR (Planner)
    # -------------------------------------------------------------
    def planner_node(state: HarnessState) -> Dict[str, Any]:
        ticket = state["ticket"]
        context = tools.get_context_summary()

        prompt = f"""Eres el Arquitecto de Software Principal de la Software Factory.
Analiza el siguiente ticket de implementación y define un plan técnico estructurado cumpliendo la arquitectura limpia (ADR-003):

TICKET ID: {ticket.ticket_id}
TÍTULO: {ticket.title}
DESCRIPCIÓN:
{ticket.description}

CRITERIOS DE ACEPTACIÓN:
{chr(10).join(f"- {c}" for c in ticket.acceptance_criteria)}

CONTEXTO ACTUAL DEL PROYECTO:
{context}

Responde ÚNICAMENTE en formato JSON válido con la siguiente estructura:
{{
  "summary": "Resumen técnico de la solución",
  "architecture_compliance": "Cómo se garantiza la separación de capas",
  "steps": ["Paso 1", "Paso 2", "..."],
  "files_to_touch": ["ruta/relativa/archivo1.ext", "ruta/relativa/archivo2.ext"]
}}
"""
        response = llm.invoke([SystemMessage(content="Eres un arquitecto de software senior."), HumanMessage(content=prompt)])
        
        # Limpieza de bloques markdown si el modelo incluye ```json
        content = response.content.strip()
        if content.startswith("```"):
            content = content.split("\n", 1)[1].rsplit("\n", 1)[0].strip()

        try:
            data = json.loads(content)
            plan = CodePlan(**data)
        except Exception:
            plan = CodePlan(
                summary=f"Implementación técnica para {ticket.ticket_id}",
                architecture_compliance="Cumplimiento con SDD y Go Clean Arch",
                steps=["Generar código fuente", "Verificar compilación"],
                files_to_touch=[]
            )

        return {"plan": plan, "status": "planned"}

    # -------------------------------------------------------------
    # NODO 2: PROGRAMADOR (Coder con Feedback de Errores)
    # -------------------------------------------------------------
    def coder_node(state: HarnessState) -> Dict[str, Any]:
        ticket = state["ticket"]
        plan = state["plan"]
        iteration = state["iteration"]
        last_error = state.get("last_test_output", "")

        feedback_section = ""
        if last_error and state["last_test_exit_code"] != 0:
            feedback_section = f"""
=============================================================
¡ALERTA DE ERROR EN LA COMPILACIÓN O PRUEBAS ANTERIORES!
El comando de prueba falló. Analiza el siguiente log de error y CORRIGE el código:
{last_error}
=============================================================
"""

        prompt = f"""Eres el Desarrollador Senior de la Software Factory.
Tu objetivo es escribir o corregir el código necesario para resolver el siguiente ticket:

TICKET ID: {ticket.ticket_id}
TÍTULO: {ticket.title}
DESCRIPCIÓN: {ticket.description}
CRITERIOS DE ACEPTACIÓN:
{chr(10).join(f"- {c}" for c in ticket.acceptance_criteria)}

PLAN TÉCNICO:
{plan.summary if plan else 'Sin plan explícito'}
{feedback_section}

Genera los archivos requeridos. Responde ÚNICAMENTE con una lista JSON de operaciones de archivo con esta estructura:
[
  {{
    "path": "ruta/relativa/al/archivo.ext",
    "action": "create",
    "content": "CODIGO COMPLETO DEL ARCHIVO AQUI",
    "explanation": "Propósito del archivo"
  }}
]
"""
        response = llm.invoke([SystemMessage(content="Eres un ingeniero de software senior en Go y TypeScript."), HumanMessage(content=prompt)])
        content = response.content.strip()
        if content.startswith("```"):
            content = content.split("\n", 1)[1].rsplit("\n", 1)[0].strip()

        try:
            raw_ops = json.loads(content)
            operations = [FileOperation(**op) for op in raw_ops]
        except Exception as e:
            # Si el modelo no dio JSON válido, registrar error para reintento
            return {
                "last_test_exit_code": 1,
                "last_test_output": f"Error parseando JSON de operaciones: {str(e)}\nRespuesta: {content[:300]}",
                "iteration": iteration + 1,
                "status": "self_correcting"
            }

        # Aplicar Guardrails
        validation = guardrails.validate_file_operations(operations)
        if not validation.is_valid:
            return {
                "last_test_exit_code": 1,
                "last_test_output": "VIOLACIÓN DE GUARDRAIL:\n" + "\n".join(validation.violations),
                "iteration": iteration + 1,
                "status": "self_correcting"
            }

        # Aplicar operaciones al disco
        applied = tools.apply_file_operations(validation.sanitized_operations)
        return {
            "applied_files": applied,
            "status": "coded",
            "pending_operations": [op.model_dump() for op in validation.sanitized_operations]
        }

    # -------------------------------------------------------------
    # NODO 3: EVALUADOR Y PRUEBAS (MCP Tester)
    # -------------------------------------------------------------
    def tester_node(state: HarnessState) -> Dict[str, Any]:
        iteration = state["iteration"] + 1
        ticket = state["ticket"]

        # Determinar comando de prueba según el tipo de archivos
        applied = state.get("applied_files", [])
        if any(f.endswith(".go") for f in applied) or (ticket and "go" in ticket.title.lower()):
            test_cmd = "go build ./..."
        elif any(f.endswith((".ts", ".tsx", ".jsx", ".js")) for f in applied):
            test_cmd = "npm test"
        else:
            test_cmd = "go test ./..."

        is_safe, reason = guardrails.validate_shell_command(test_cmd)
        if not is_safe:
            return {
                "iteration": iteration,
                "last_test_exit_code": 1,
                "last_test_output": reason,
                "status": "self_correcting"
            }

        exit_code, output = tools.run_shell(test_cmd)
        compact_output = tools.truncate_error_log(output, max_lines=20)
        return {
            "iteration": iteration,
            "test_command": test_cmd,
            "last_test_exit_code": exit_code,
            "last_test_output": compact_output,
            "status": "tested" if exit_code == 0 else "self_correcting"
        }

    # -------------------------------------------------------------
    # NODO 4: COMMITTER (Versionado Automático)
    # -------------------------------------------------------------
    def committer_node(state: HarnessState) -> Dict[str, Any]:
        ticket = state["ticket"]
        branch_name = ticket.branch or f"feature/{ticket.ticket_id.lower()}"
        
        # Git branch y commit
        tools.run_shell(f"git checkout -B {branch_name}")
        tools.run_shell("git add .")
        tools.run_shell(f'git commit -m "feat({ticket.ticket_id}): {ticket.title}"')

        return {
            "status": "completed",
            "git_branch": branch_name
        }

    # -------------------------------------------------------------
    # ENLACES CONDICIONALES (Decisión de Autocorrección)
    # -------------------------------------------------------------
    def should_continue(state: HarnessState) -> str:
        if state["last_test_exit_code"] == 0:
            return "committer"
        if state["iteration"] >= state["max_iterations"]:
            return "failed_end"
        return "coder"

    # -------------------------------------------------------------
    # ENSAMBLAR EL GRAFO
    # -------------------------------------------------------------
    workflow = StateGraph(HarnessState)

    workflow.add_node("planner", planner_node)
    workflow.add_node("coder", coder_node)
    workflow.add_node("tester", tester_node)
    workflow.add_node("committer", committer_node)

    workflow.set_entry_point("planner")
    workflow.add_edge("planner", "coder")
    workflow.add_edge("coder", "tester")

    workflow.add_conditional_edges(
        "tester",
        should_continue,
        {
            "committer": "committer",
            "coder": "coder",
            "failed_end": END
        }
    )
    workflow.add_edge("committer", END)

    return workflow.compile()
