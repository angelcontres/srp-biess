"""
Pruebas de Verificación de Componentes del Agent Harness.
"""

from pathlib import Path
from harness.models import FileOperation
from harness.guardrails import HarnessGuardrails
from harness.tools import HarnessToolchain

ROOT = Path(__file__).resolve().parent.parent

def test_ticket_parser():
    toolchain = HarnessToolchain(ROOT)
    ticket = toolchain.parse_ticket("SRP-001")
    assert ticket is not None, "El ticket SRP-001 debe existir en Docs/TICKETS.md"
    assert ticket.ticket_id == "SRP-001"
    assert "backend-clean-arch" in ticket.branch
    assert len(ticket.acceptance_criteria) >= 2
    print("[OK] Ticket Parser: Exito")

def test_guardrails_path_traversal():
    guardrails = HarnessGuardrails(ROOT)
    op = FileOperation(
        path="../evil.py",
        action="create",
        content="print('hack')",
        explanation="test"
    )
    res = guardrails.validate_file_operations([op])
    assert not res.is_valid, "Debe rechazar rutas que escapen de la raíz"
    print("[OK] Guardrail Path Traversal: Exito")

def test_guardrails_clean_architecture():
    guardrails = HarnessGuardrails(ROOT)
    op = FileOperation(
        path="internal/domain/user.go",
        action="create",
        content='package domain\nimport "database/sql"\ntype User struct{}',
        explanation="test"
    )
    res = guardrails.validate_file_operations([op])
    assert not res.is_valid, "Debe rechazar imports de base de datos en el dominio (ADR-003)"
    print("[OK] Guardrail Clean Architecture (ADR-003): Exito")

def test_guardrails_shell_command():
    guardrails = HarnessGuardrails(ROOT)
    ok, _ = guardrails.validate_shell_command("go build ./cmd/api")
    assert ok, "go build debe ser un comando permitido"
    
    bad, _ = guardrails.validate_shell_command("format C: /y")
    assert not bad, "Comandos destructivos deben ser bloqueados"
    print("[OK] Guardrail Shell Command Allowlist: Exito")

if __name__ == "__main__":
    test_ticket_parser()
    test_guardrails_path_traversal()
    test_guardrails_clean_architecture()
    test_guardrails_shell_command()
    print("\nTODOS LOS COMPONENTES DEL ARNES FUNCIONAN PERFECTAMENTE.")
