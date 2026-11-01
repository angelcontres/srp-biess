"""
Punto de Entrada CLI del Agent Harness (Software Factory).
Permite listar tickets, ejecutar agentes con auto-corrección y monitorear el progreso en tiempo real.
"""

import sys
import argparse
from pathlib import Path

from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.markdown import Markdown

from harness.tools import HarnessToolchain
from harness.graph import create_software_factory_graph
from harness.models import HarnessState

console = Console()
ROOT_DIR = Path(__file__).resolve().parent.parent


def cmd_list_tickets(args):
    """Muestra una tabla con todos los tickets disponibles en Docs/TICKETS.md."""
    toolchain = HarnessToolchain(ROOT_DIR)
    tickets_path = ROOT_DIR / "docs" / "TICKETS.md"
    if not tickets_path.exists():
        tickets_path = ROOT_DIR / "Docs" / "TICKETS.md"

    if not tickets_path.exists():
        console.print("[bold red]No se encontró el archivo Docs/TICKETS.md[/bold red]")
        return

    content = tickets_path.read_text(encoding="utf-8")
    import re
    ticket_ids = re.findall(r"### Ticket `(SRP-\d+)`", content)

    table = Table(title="Catálogo de Tickets — Software Factory (SRP-BIESS)", show_header=True, header_style="bold cyan")
    table.add_column("Ticket ID", style="bold yellow", width=12)
    table.add_column("Título", style="white")
    table.add_column("Rama Asignada", style="green")
    table.add_column("Responsable", style="magenta")

    for tid in ticket_ids:
        info = toolchain.parse_ticket(tid)
        if info:
            table.add_row(info.ticket_id, info.title, info.branch, info.assigned_to)

    console.print(table)


def cmd_run_ticket(args):
    """Ejecuta el ciclo autónomo de LangGraph para resolver el ticket especificado."""
    ticket_id = args.ticket_id.upper()
    toolchain = HarnessToolchain(ROOT_DIR)
    ticket = toolchain.parse_ticket(ticket_id)

    if not ticket:
        console.print(f"[bold red]Error: No se encontró el ticket '{ticket_id}' en Docs/TICKETS.md[/bold red]")
        return

    console.print(Panel.fit(
        f"[bold cyan]EJECUTANDO AGENT HARNESS (SOFTWARE FACTORY)[/bold cyan]\n"
        f"[yellow]Ticket:[/yellow] {ticket.ticket_id} — {ticket.title}\n"
        f"[green]Rama:[/green] {ticket.branch}\n"
        f"[magenta]Responsable:[/magenta] {ticket.assigned_to}\n"
        f"[blue]Criterios de Aceptación:[/blue] {len(ticket.acceptance_criteria)} reglas",
        border_style="cyan"
    ))

    # Inicializar Estado de LangGraph
    initial_state: HarnessState = {
        "ticket_id": ticket.ticket_id,
        "ticket": ticket,
        "plan": None,
        "pending_operations": [],
        "applied_files": [],
        "iteration": 0,
        "max_iterations": args.max_iterations,
        "test_command": "",
        "last_test_output": "",
        "last_test_exit_code": -1,
        "status": "init",
        "git_branch": ticket.branch,
        "error_message": None
    }

    try:
        graph = create_software_factory_graph(ROOT_DIR, provider=args.provider, model=args.model)
    except Exception as e:
        console.print(f"[bold red]Error al inicializar el modelo o grafo:[/bold red] {e}")
        console.print("[yellow]Verifica que hayas configurado tu archivo harness/.env con tus API keys.[/yellow]")
        return

    console.print("[bold blue]Iniciando flujo cíclico en LangGraph...[/bold blue]")
    
    final_state = graph.invoke(initial_state)

    if final_state["status"] == "completed":
        console.print(Panel.fit(
            f"[bold green]¡TICKET COMPLETADO CON ÉXITO![/bold green]\n"
            f"[white]Archivos aplicados:[/white] {', '.join(final_state.get('applied_files', []))}\n"
            f"[white]Rama Git creada:[/white] {final_state.get('git_branch')}\n"
            f"[white]Iteraciones necesarias:[/white] {final_state.get('iteration', 1)}",
            border_style="green"
        ))
    else:
        console.print(Panel.fit(
            f"[bold red]EL PROCESO SE DETUVO SIN ÉXITO[/bold red]\n"
            f"[white]Iteraciones ejecutadas:[/white] {final_state.get('iteration')}\n"
            f"[white]Último error:[/white] {final_state.get('last_test_output')[:400]}",
            border_style="red"
        ))


def main():
    parser = argparse.ArgumentParser(description="Agent Harness & AI Software Factory CLI")
    subparsers = parser.add_subparsers(dest="command")

    # Comando: list
    subparsers.add_parser("list", help="Listar todos los tickets del backlog")

    # Comando: run
    run_parser = subparsers.add_parser("run", help="Ejecutar el arnés sobre un ticket")
    run_parser.add_argument("ticket_id", type=str, help="ID del ticket (ej. SRP-001)")
    run_parser.add_argument("--provider", type=str, default=None, help="Proveedor de IA (gemini, openai, deepseek, ollama)")
    run_parser.add_argument("--model", type=str, default=None, help="Nombre del modelo específico")
    run_parser.add_argument("--max-iterations", type=int, default=5, help="Máximo número de intentos de autocorrección")

    args = parser.parse_args()

    if args.command == "list":
        cmd_list_tickets(args)
    elif args.command == "run":
        cmd_run_ticket(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
