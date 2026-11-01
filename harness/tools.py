"""
Protocolo de Conectividad y Herramientas (MCP / Tool Execution Layer).
Interactúa con el sistema de archivos, el catálogo de tickets y la terminal de desarrollo.
"""

import os
import re
import subprocess
from pathlib import Path
from typing import Optional, List, Tuple
from harness.models import TicketInfo, FileOperation


class HarnessToolchain:
    """Implementación de herramientas desacopladas compatibles con el estándar MCP."""

    def __init__(self, root_dir: Path):
        self.root_dir = root_dir.resolve()

    def parse_ticket(self, ticket_id: str) -> Optional[TicketInfo]:
        """Extrae la información de un ticket específico desde Docs/TICKETS.md."""
        tickets_path = self.root_dir / "docs" / "TICKETS.md"
        if not tickets_path.exists():
            tickets_path = self.root_dir / "Docs" / "TICKETS.md"
            if not tickets_path.exists():
                return None

        content = tickets_path.read_text(encoding="utf-8")
        
        # Buscar la sección del ticket
        pattern = rf"### Ticket `({ticket_id})`:\s*(.+?)\n(.*?)(?=\n### Ticket |\Z)"
        match = re.search(pattern, content, re.DOTALL)
        if not match:
            return None

        t_id = match.group(1).strip()
        title = match.group(2).strip()
        body = match.group(3).strip()

        # Extraer campos
        t_type = "Feature"
        m_type = re.search(r"\*\s*\*\*Tipo:\*\*\s*(.+)", body)
        if m_type:
            t_type = m_type.group(1).strip()

        assigned_to = ""
        m_assign = re.search(r"\*\s*\*\*Asignado:\*\*\s*(.+)", body)
        if m_assign:
            assigned_to = m_assign.group(1).strip()

        branch = f"feature/{t_id.lower()}"
        m_branch = re.search(r"\*\s*\*\*Rama:\*\*\s*`?([^`\n]+)`?", body)
        if m_branch:
            branch = m_branch.group(1).strip()

        description = ""
        m_desc = re.search(r"\*\s*\*\*Descripción:\*\*\s*(.+?)(?=\n\*\s*\*\*Criterios de Aceptación:|\Z)", body, re.DOTALL)
        if m_desc:
            description = m_desc.group(1).strip()

        criteria = []
        m_crit = re.search(r"\*\s*\*\*Criterios de Aceptación:\*\*\s*(.+)", body, re.DOTALL)
        if m_crit:
            raw_crit = m_crit.group(1)
            for line in raw_crit.splitlines():
                line = line.strip()
                if line.startswith("- [ ]") or line.startswith("- [x]") or line.startswith("*"):
                    crit_text = re.sub(r"^[-*]\s*(\[[ x]\]\s*)?", "", line).strip()
                    if crit_text:
                        criteria.append(crit_text)

        return TicketInfo(
            ticket_id=t_id,
            title=title,
            ticket_type=t_type,
            assigned_to=assigned_to,
            branch=branch,
            description=description,
            acceptance_criteria=criteria
        )

    def apply_file_operations(self, operations: List[FileOperation]) -> List[str]:
        """Aplica las operaciones de archivo aprobadas sobre el workspace."""
        applied = []
        for op in operations:
            clean_path = op.path.replace("\\", "/").strip().lstrip("/")
            target_path = self.root_dir / clean_path

            if op.action in ("create", "modify"):
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text(op.content, encoding="utf-8")
                applied.append(clean_path)
            elif op.action == "delete":
                if target_path.exists():
                    target_path.unlink()
                    applied.append(f"DELETED: {clean_path}")

        return applied

    def run_shell(self, command: str, timeout_seconds: int = 60) -> Tuple[int, str]:
        """Ejecuta un comando en la terminal local y retorna exit code y salida completa."""
        try:
            result = subprocess.run(
                command,
                cwd=str(self.root_dir),
                shell=True,
                capture_output=True,
                text=True,
                timeout=timeout_seconds,
                encoding="utf-8",
                errors="replace"
            )
            output = (result.stdout + "\n" + result.stderr).strip()
            return result.returncode, output
        except subprocess.TimeoutExpired:
            return 124, f"TIMEOUT: El comando '{command}' excedió el límite de {timeout_seconds} segundos."
        except Exception as e:
            return 1, f"ERROR AL EJECUTAR COMANDO: {str(e)}"

    def truncate_error_log(self, raw_output: str, max_lines: int = 25) -> str:
        """
        [AHORRO DE TOKENS]: Compacta los logs de compilación/test manteniendo
        únicamente las líneas de error relevantes y eliminando ruido innecesario.
        """
        lines = [line.strip() for line in raw_output.splitlines() if line.strip()]
        if len(lines) <= max_lines:
            return "\n".join(lines)

        # Filtrar líneas críticas que contengan indicadores de error típicos de Go, Node y GCC
        error_lines = [
            l for l in lines 
            if any(k in l.lower() for k in ("error", "fail", "undefined", "cannot", "panic", "syntax", "fatal"))
        ]
        
        if error_lines:
            # Retornar las líneas de error más las últimas 10 líneas de contexto
            tail = lines[-10:]
            selected = list(dict.fromkeys(error_lines[:15] + tail))
            return "\n".join(selected[:max_lines])
            
        return "\n".join(lines[-max_lines:])

    def get_context_summary(self) -> str:
        """[AHORRO DE TOKENS]: Obtiene solo la estructura mínima indispensable."""
        summary = ["ESTRUCTURA CLAVE:"]
        for root, dirs, files in os.walk(self.root_dir):
            dirs[:] = [d for d in dirs if d not in (".git", "node_modules", ".venv", "__pycache__", "dist", "bin", "docs", "Docs")]
            rel_dir = os.path.relpath(root, self.root_dir)
            if rel_dir == ".":
                rel_dir = ""
            for f in files:
                if f.endswith((".go", ".mod", ".sql", ".json")):
                    rel_file = os.path.join(rel_dir, f).replace("\\", "/")
                    summary.append(f"- {rel_file}")
        return "\n".join(summary[:30])
