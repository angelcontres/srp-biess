import re
import subprocess
from pathlib import Path

def parse_tickets():
    content = Path("docs/TICKETS.md").read_text(encoding="utf-8")
    
    # Each ticket starts with: ### Ticket `SRP-XXX`: Title
    pattern = r"### Ticket `(SRP-\d+)`:\s*([^\n\r]+)(.*?)(?=\n### Ticket |\Z)"
    matches = re.findall(pattern, content, re.DOTALL)
    
    parsed = []
    for tid, title, body in matches:
        title = title.strip()
        body = body.strip()
        # Remove trailing markdown dividers ---
        body = re.sub(r"\n---\s*$", "", body).strip()
        
        # Extract metadata
        m_type = re.search(r"\*\s*\*\*Tipo:\*\*\s*(.+)", body)
        t_type = m_type.group(1).strip() if m_type else "Feature"
        
        m_assign = re.search(r"\*\s*\*\*Asignado:\*\*\s*(.+)", body)
        assigned = m_assign.group(1).strip() if m_assign else ""
        
        m_branch = re.search(r"\*\s*\*\*Rama:\*\*\s*`?([^`\n]+)`?", body)
        branch = m_branch.group(1).strip() if m_branch else ""
        
        parsed.append({
            "id": tid,
            "title": f"[{tid}] {title}",
            "type": t_type,
            "assigned": assigned,
            "branch": branch,
            "body": f"### Metadatos\n* **Ticket:** `{tid}`\n* **Tipo:** {t_type}\n* **Asignado:** {assigned}\n* **Rama sugerida:** `{branch}`\n\n---\n\n{body}"
        })
    return parsed

if __name__ == "__main__":
    tickets = parse_tickets()
    print(f"Total tickets parsed: {len(tickets)}")
    for t in tickets:
        print(f"{t['id']} | {t['title']} | {t['assigned']}")
