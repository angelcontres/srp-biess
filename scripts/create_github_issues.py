import re
import subprocess
from pathlib import Path

def get_tickets():
    content = Path("docs/TICKETS.md").read_text(encoding="utf-8")
    pattern = r"### Ticket `(SRP-\d+)`:\s*([^\n\r]+)(.*?)(?=\n### Ticket |\Z)"
    matches = re.findall(pattern, content, re.DOTALL)
    
    tickets = []
    for tid, title, body in matches:
        title = title.strip()
        body = body.strip()
        body = re.sub(r"\n---\s*$", "", body).strip()
        
        # Clean up LaTeX or special chars for markdown readability if needed
        # but keep markdown formatting
        m_type = re.search(r"\*\s*\*\*Tipo:\*\*\s*(.+)", body)
        t_type = m_type.group(1).strip() if m_type else "Feature"
        
        m_assign = re.search(r"\*\s*\*\*Asignado:\*\*\s*(.+)", body)
        assigned = m_assign.group(1).strip() if m_assign else ""
        
        m_branch = re.search(r"\*\s*\*\*Rama:\*\*\s*`?([^`\n]+)`?", body)
        branch = m_branch.group(1).strip() if m_branch else ""
        
        # Determine labels
        labels = ["enhancement"]
        if "SRP-001" in tid or "SRP-002" in tid or "SRP-003" in tid:
            labels.extend(["backend", "go"])
        elif "SRP-004" in tid or "SRP-005" in tid:
            labels.extend(["backend", "database", "postgres"])
        elif "SRP-006" in tid or "SRP-007" in tid:
            labels.extend(["backend", "cache", "redis"])
        elif "SRP-008" in tid or "SRP-009" in tid or "SRP-010" in tid:
            labels.extend(["backend", "integrations"])
        elif "SRP-011" in tid or "SRP-012" in tid or "SRP-013" in tid or "SRP-014" in tid or "SRP-015" in tid or "SRP-016" in tid:
            labels.extend(["frontend", "react"])

        issue_body = f"""## Metadatos del Ticket
* **ID:** `{tid}`
* **Tipo:** {t_type}
* **Asignado:** **{assigned}**
* **Rama Git Asignada:** `{branch}`

---

{body}
"""
        tickets.append({
            "id": tid,
            "title": f"[{tid}] {title}",
            "body": issue_body,
            "labels": labels
        })
    return tickets

def create_issues():
    tickets = get_tickets()
    print(f"Creating {len(tickets)} issues on GitHub repo angelcontres/srp-biess...")
    
    # Ensure custom labels exist or create them
    needed_labels = [
        ("backend", "1d76db", "Backend Go Fiber"),
        ("frontend", "0e8a16", "Frontend React Vite"),
        ("go", "00add8", "Lenguaje Go"),
        ("react", "61dafb", "React SPA"),
        ("database", "5319e7", "PostgreSQL DDL & Transacciones"),
        ("postgres", "336791", "PostgreSQL"),
        ("cache", "d93f0b", "Redis Cache & Lock"),
        ("redis", "dc382d", "Redis"),
        ("integrations", "fbca04", "Integraciones Mocks & Pagos")
    ]
    
    for name, color, desc in needed_labels:
        subprocess.run(
            ["gh", "label", "create", name, "--color", color, "--description", desc, "--force"],
            capture_output=True
        )

    created_count = 0
    for t in tickets:
        cmd = [
            "gh", "issue", "create",
            "--repo", "angelcontres/srp-biess",
            "--title", t["title"],
            "--body", t["body"]
        ]
        for l in t["labels"]:
            cmd.extend(["--label", l])
            
        res = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8")
        if res.returncode == 0:
            url = res.stdout.strip()
            print(f"[OK] {t['id']} -> {url}")
            created_count += 1
        else:
            print(f"[ERROR] {t['id']}: {res.stderr.strip()}")
            
    print(f"\nDone! Successfully created {created_count}/{len(tickets)} GitHub Issues.")

if __name__ == "__main__":
    create_issues()
