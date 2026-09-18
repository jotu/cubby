# .agents/skills/

Modular skill definitions for coding agents. Tool-agnostic — works with
OpenCode, Cursor, Copilot, Gemini CLI, and any agent that reads `.agents/skills/`.

Each subdirectory contains a `SKILL.md` with domain-specific guidance that agents
load on demand. See `AGENTS.md` at repo root for project-wide rules.

## Skills

| Skill | Description |
| --- | --- |
| `conventional-commits` | Commit message format, PR titles, scopes |
| `docker` | Distroless builds, layered caching, SQLite volumes |
| `flyme` | Safe, repeatable SQLite data migration workflow across dev and prod |
| `go-backend` | net/http 1.22+ routing, handlers, middleware, JSON patterns |
| `go-tdd` | Table-driven tests, httptest, t.Helper, TDD workflow |
| `frontend-material-ui` | Expert UI design with Material Design 3, Lucide icons, Google Fonts, and light/dark themes |
| `mise` | Task runner, tool management, file-based tasks |
| `rest-api` | URL design, status codes, error format, pagination |
| `sqlite-patterns` | WAL mode, migrations, FTS5, dynamic updates |
