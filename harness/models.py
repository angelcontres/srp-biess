"""
Modelos de Datos y Estado para el Arnés de Agentes (Pydantic & LangGraph).
"""

from typing import List, Optional, Literal, Dict, Any
from typing_extensions import TypedDict
from pydantic import BaseModel, Field


class TicketInfo(BaseModel):
    """Información estructurada extraída del catálogo Docs/TICKETS.md."""
    ticket_id: str = Field(description="Identificador del ticket, ej. SRP-001")
    title: str = Field(description="Título del ticket")
    ticket_type: str = Field(default="Feature", description="Tipo de tarea")
    assigned_to: str = Field(default="", description="Miembro del equipo asignado")
    branch: str = Field(description="Nombre de la rama Git a crear")
    description: str = Field(description="Descripción detallada de la tarea")
    acceptance_criteria: List[str] = Field(default_factory=list, description="Criterios de aceptación a validar")


class FileOperation(BaseModel):
    """Operación sobre el sistema de archivos generada por el agente Coder."""
    path: str = Field(description="Ruta relativa del archivo respecto a la raíz del proyecto")
    action: Literal["create", "modify", "delete"] = Field(description="Acción a realizar")
    content: str = Field(description="Contenido completo del archivo para create/modify")
    explanation: str = Field(description="Breve justificación de este cambio")


class CodePlan(BaseModel):
    """Plan de implementación estructurado generado por el agente Arquitecto/Planner."""
    summary: str = Field(description="Resumen del enfoque técnico")
    architecture_compliance: str = Field(description="Cómo cumple con las reglas de arquitectura de Docs/SDD")
    steps: List[str] = Field(description="Pasos secuenciales de implementación")
    files_to_touch: List[str] = Field(description="Lista de rutas de archivos que se crearán o modificarán")


class GuardrailValidation(BaseModel):
    """Resultado de la validación de seguridad y arquitectura previa a la ejecución."""
    is_valid: bool = Field(description="True si pasa todos los guardrails")
    violations: List[str] = Field(default_factory=list, description="Lista de violaciones detectadas si is_valid es False")
    sanitized_operations: List[FileOperation] = Field(default_factory=list, description="Operaciones seguras aprobadas")


class HarnessState(TypedDict):
    """
    Estado compartido que fluye a través del Grafo de Estados cíclico de LangGraph.
    """
    ticket_id: str
    ticket: Optional[TicketInfo]
    plan: Optional[CodePlan]
    pending_operations: List[Dict[str, Any]]
    applied_files: List[str]
    iteration: int
    max_iterations: int
    test_command: str
    last_test_output: str
    last_test_exit_code: int
    status: Literal["init", "planned", "coded", "tested", "self_correcting", "completed", "failed"]
    git_branch: str
    error_message: Optional[str]
