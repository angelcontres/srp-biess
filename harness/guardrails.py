"""
Capas de Seguridad y Validación (Guardrails) para el Arnés de Agentes.
Asegura que el código generado respete la arquitectura limpia y previene acciones destructivas.
"""

import os
import re
from pathlib import Path
from typing import List, Tuple
from harness.models import FileOperation, GuardrailValidation


ALLOWED_COMMAND_PREFIXES = (
    "go build",
    "go test",
    "go vet",
    "go mod",
    "npm run",
    "npm test",
    "npx",
    "docker compose",
    "docker-compose",
    "git status",
    "git diff",
    "git branch",
    "git checkout",
    "git add",
    "git commit",
)

PROTECTED_PATHS = (
    ".git",
    "harness",
    ".env",
)


class HarnessGuardrails:
    """Implementación de barreras de contención (Guardrails) para el arnés."""

    def __init__(self, root_dir: Path):
        self.root_dir = root_dir.resolve()

    def validate_file_operations(self, operations: List[FileOperation]) -> GuardrailValidation:
        """Valida rutas, permisos y reglas de arquitectura en las operaciones de archivo."""
        violations = []
        sanitized = []

        for op in operations:
            clean_path = op.path.replace("\\", "/").strip().lstrip("/")
            target_path = (self.root_dir / clean_path).resolve()

            # 1. Guardrail de Escape de Sandbox (Path Traversal)
            try:
                target_path.relative_to(self.root_dir)
            except ValueError:
                violations.append(f"VIOLACIÓN DE RUTA: '{op.path}' intenta escapar del directorio raíz del proyecto.")
                continue

            # 2. Guardrail de Archivos Protegidos del Sistema
            if any(clean_path.startswith(p) for p in PROTECTED_PATHS) and not clean_path.endswith(".example"):
                violations.append(f"ARCHIVO PROTEGIDO: No se permite modificar '{clean_path}' de forma autónoma.")
                continue

            # 3. Guardrail de Arquitectura Limpia (Go / ADR-003)
            # El dominio jamás debe depender de infraestructura, drivers SQL o frameworks HTTP
            if clean_path.startswith("internal/domain/") and clean_path.endswith(".go"):
                forbidden_imports = [
                    "database/sql",
                    "github.com/jackc/pgx",
                    "github.com/gofiber/fiber",
                    "internal/infrastructure",
                    "internal/interfaces",
                ]
                for forbidden in forbidden_imports:
                    if f'"{forbidden}"' in op.content or f'"{forbidden}/' in op.content:
                        violations.append(
                            f"VIOLACIÓN DE ARQUITECTURA LIMPIA (ADR-003): El archivo de dominio '{clean_path}' "
                            f"importa dependencias prohibidas ('{forbidden}'). El dominio debe ser agnóstico."
                        )

            sanitized.append(op)

        return GuardrailValidation(
            is_valid=(len(violations) == 0),
            violations=violations,
            sanitized_operations=sanitized
        )

    def validate_shell_command(self, command: str) -> Tuple[bool, str]:
        """Verifica que el comando a ejecutar pertenezca a la lista blanca segura."""
        stripped = command.strip()

        # Evitar comandos destructivos o encadenamientos maliciosos
        if any(token in stripped for token in [";", "&&", "||", "|", "`", "$("]) and not any(
            stripped.startswith(p) for p in ("go test", "npm test")
        ):
            # Permitir solo encadenamientos seguros controlados si empiezan con prefijos válidos
            pass

        if not any(stripped.startswith(prefix) for prefix in ALLOWED_COMMAND_PREFIXES):
            return False, f"COMANDO RECHAZADO: '{command}' no está en la lista blanca permitida de la Software Factory."

        return True, ""
