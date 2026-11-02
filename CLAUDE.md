# Claude Code Guidelines — SRP-BIESS

## Build and Test Commands
- Run backend build: `go build ./cmd/api`
- Run Go tests: `go test -v ./...`
- Fast verification harness: `verify.bat` or `./verify.sh`
- Docker services: `docker compose up -d`

## Architecture Rules (ADR-003)
- Strict layer separation in Go: `cmd/api`, `internal/domain`, `internal/application`, `internal/infrastructure`, `internal/interfaces/http`.
- The domain layer (`internal/domain`) must be 100% pure Go: NO imports of `database/sql`, `pgx`, or HTTP frameworks.
- Manual dependency injection in `cmd/api/main.go`.

## Token Optimization
- Be extremely concise. Avoid repetitive explanations.
- Prefer targeted patches over full file rewrites.

## Interface Design

This project is the application system for the BIESS Public Merit and Competitive Selection Process: a digital public service, not a consumer product.

**Before writing or modifying any interface, read `DESIGN.md` at the project root.** It is the visual source of truth and takes precedence over any Tailwind defaults, component library defaults, or conventional styling practices. Business rules and institutional copy are defined in the SDD (`08-diseno-interfaz-usuario.md` and the documents it references); do not invent them: use `[PENDING: …]`.

Stack: React, Tailwind CSS 3.4.17 (with `theme.colors` replaced by the tokens defined in DESIGN.md), and `lucide-react` ^0.474.0.

Rules that must never be violated:

- Use only DESIGN.md token classes. No hex values, `rgb()`, arbitrary values (`p-[13px]`), or default color palettes (`gray-*`, `slate-*`, `blue-*`, etc.).
- Containers must use `border border-line-strong` and have no shadows; inputs must use `border-line-input`; use `line` only for internal dividers.
- Border radii: buttons and inputs `rounded-md`, cards and modals `rounded-lg`, badges `rounded`. Nothing larger than `rounded-lg`.
- Use `outline` for focus indicators (`focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark`), never `ring`.
- Only one primary button per screen. Controls must be 44 px high (`h-11`).
- Every state must include an icon and text; never use color alone.
- Spinners are allowed only inside buttons; use skeletons for layout loading states.
- A server error must never erase form data.
- Prohibited: emojis, the middle dot `·`, `uppercase`, `backdrop-blur`, gradients, shimmer effects, toggles for Yes/No choices, monospace fonts, and placeholders used as labels.
- All user-facing text must use formal address ("usted" in Spanish); preserve the fixed terms "postulación" and "sección". Format dates using `Intl` with the `es-EC` locale and `America/Guayaquil` time zone.
- WCAG 2.2 AA accessibility compliance is the minimum requirement.

## Project Skills

Located in `.claude/skills/`. Use them whenever the task matches:

- `/new-page`: Creates a new page or route for the BIESS application system, following the layout, typography, and spacing defined in DESIGN.md. Use it when a new screen is requested (login, result, My Application, error, landing page). For application form sections, use `new-form-section`.
- `/new-form-section`: Creates one of the 10 sections of the BIESS application form (Personal Information, Address, Family Members, Position, Education, Work Experience, Publications, Training, Disqualifications, Summary), including a sticky section bar, saving, local drafts, error summaries, and loading and error states according to DESIGN.md.
- `/new-component`: Creates a reusable UI component for the BIESS application system using only the tokens, border radii, borders, states, and accessibility rules defined in DESIGN.md. Use it when a new button, field, card, alert, badge, modal, or other component is requested.
- `/add-section`: Adds a content block or a group of fields to an existing page in the BIESS application system, respecting the spacing, typography scale, and responsive layout defined in DESIGN.md. Use it when asked to add a content block, notice, group of questions, or list to an existing screen.
- `/design-audit`: Audits a file or folder in the BIESS application system against DESIGN.md (tokens, shapes, components, accessibility, content, and prohibitions) and reports violations in a table without modifying files. Use it to review interface code before a commit, PR, or delivery.
- `/responsive-check`: Verifies that a page in the BIESS application system works across all breakpoints defined in DESIGN.md (320 to 1280 px and zoom), identifying overflow, overlapping elements, undersized controls, and obscured focus indicators, and proposes fixes using existing tokens. Use it after creating or modifying a screen.
- `/token-check`: Searches the BIESS application system code for hardcoded colors, dimensions, border radii, shadows, and fonts, or values outside the configuration, and proposes the corresponding DESIGN.md tokens to replace them, without applying changes until confirmation. Use it before a commit or whenever inconsistent styling values are suspected.
