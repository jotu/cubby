# AGENTS.md — Cubby

> Home inventory system. Go backend, frontend TBD. SQLite storage.
> Modular skills live in `.agents/skills/`. This file is the root reference.

## Architecture

```
backend/
  cmd/server/          # Entry point
  internal/
    api/               # HTTP handlers (net/http 1.22+)
    db/                # SQLite queries, migrations
    models/            # Domain types, request/response structs
    qr/                # QR code generation
    service/           # Business logic layer
frontend/              # TBD — will be Node 22 / npm
```

Module: `github.com/joacim/cubby` — Go 1.25, zero-dep except `modernc.org/sqlite`.

## Commands (mise)

All tooling managed by `mise.toml`. All tasks are bash scripts in `.mise/tasks/`.

```sh
mise run dev                    # backend + frontend
mise run dev:backend            # go run ./cmd/server
mise run dev:frontend           # npm run dev

mise run build                  # both
mise run build:backend          # -> bin/cubby
mise run build:pi               # cross-compile ARM64

mise run test                   # go test ./...
mise run lint                   # golangci-lint
mise run fmt                    # gofumpt + prettier
mise run fmt-check              # check without writing
mise run vet                    # go vet

mise run verify                 # guardrail: fmt-check + vet + lint + test
mise run clean                  # remove bin/ and dist/

mise run flyme:up               # apply migrations to CUBBY_DB_PATH
mise run flyme:status           # list migration states
mise run flyme:check            # validate migration files (read-only)
mise run flyme:down -- 1        # roll back latest migration
mise run flyme:redo             # down + up for latest migration
mise run flyme:create -- "<description>"  # create new migration files

mise run docker:build           # docker compose build
mise run docker:up              # docker compose up -d
mise run docker:down            # docker compose down

mise run db:backup [output-path]             # backup SQLite database
mise run db:restore -- --yes --stopped <backup-file>   # restore backup (app must be stopped)
```

### Running specific tests

```sh
cd backend && go test ./internal/db/ -run TestCreateItem -v   # single test
cd backend && go test -race ./...                              # race detector
cd backend && go test -coverprofile=coverage.out ./...         # coverage
```

## Workflow: TDD + Tidy

Every change follows **Red → Green → Tidy**. No exceptions.
See `.agents/skills/go-tdd/SKILL.md` for patterns and examples.

1. **Red** — Write a failing test. The test defines the contract.
2. **Green** — Write the minimum code to pass. Stop.
3. **Tidy** — Refactor only. Tests must still pass. Run `mise run verify`.

Tidy = remove duplication, improve names, extract functions. Don't add features.
Don't refactor and fix bugs in the same step.

## Code Style — Go (summary)

Full patterns in `.agents/skills/go-backend/SKILL.md`.

- **Imports**: stdlib first, blank line, then external
- **Errors**: always `fmt.Errorf("context: %w", err)`, lowercase context
- **Naming**: `PascalCase` exported, `camelCase` unexported, `snake_case` SQL columns
- **JSON tags**: `camelCase` — `json:"locationId"`, `json:"createdAt"`
- **Structs**: domain models in `internal/models/models.go`, pointer types for nullable/optional
- **Handlers**: Go 1.22+ method routing `"GET /v1/items/{id}"`, `r.PathValue("id")`
- **Database**: raw SQL, `*sql.DB` injected, WAL mode, FTS5 — see `sqlite-patterns` skill
- **Tests**: `*_test.go` next to source, table-driven, in-memory SQLite — see `go-tdd` skill
- **Logging**: `log.Printf` from stdlib

## Code Style — Frontend (TBD)

Node 22, npm. Build output: `frontend/dist/` → served as static files by Go backend.

## Files to Never Commit

`*.db`, `*.db-journal`, `*.db-wal`, `.env`, `.env.local`, `bin/`, `frontend/node_modules/`, `frontend/dist/`

## Agent Rules

- **No external libraries** unless thoroughly justified with alternatives. Always prefer stdlib. The only approved external dep is `modernc.org/sqlite`. Before adding any dependency, explain why it's needed and propose at least two alternatives.
- **Read before write**: always read the file/function before editing
- **Run `mise run verify`** before committing (fmt-check + vet + lint + test)
- **Conventional commits**: see `.agents/skills/conventional-commits/SKILL.md`
- **Match existing patterns**: when in doubt, grep for similar code
- **No `any` type**: use concrete types. No `interface{}` unless truly generic
- **No TODO comments without issue**: if something needs doing, do it or file an issue

## Skills

| Skill | When to use |
| --- | --- |
| `conventional-commits` | Writing commit messages and PR titles |
| `docker` | Building containers, Dockerfile changes |
| `frontend-material-ui` | Frontend expert UI design with Material Design 3, Lucide icons, Google Fonts, and light/dark theme tokens |
| `go-backend` | HTTP handlers, middleware, JSON, service layer |
| `go-tdd` | Writing and structuring Go tests |
| `mise` | Adding tasks, managing tools |
| `rest-api` | Designing endpoints, status codes, error format |
| `sqlite-patterns` | Queries, migrations, FTS5, transactions |
