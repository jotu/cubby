---
name: mise
description: Mise tooling and task runner patterns — tool versioning, file-based tasks, environment management
---

## When to use

Use when adding tools, tasks, or environment configuration to the project. Mise manages all tooling (no Homebrew/apt for dev tools) and replaces Make as the task runner.

## Tool management

All dev tools declared in `mise.toml` at project root:

```toml
[tools]
go = "1.25"
node = "22"
"golangci-lint" = "latest"
gofumpt = "latest"
"npm:prettier" = "latest"
```

- Pin major versions for languages (`go = "1.25"`, `node = "22"`)
- Use `latest` for tools that are backwards-compatible (`golangci-lint`, `gofumpt`)
- Prefix npm packages with `npm:` (`"npm:prettier" = "latest"`)

### Commands

```sh
mise install          # install all tools
mise ls               # list installed versions
mise current          # show active versions
mise use go@1.26      # update a tool version
```

## File-based tasks

Tasks live as executable bash scripts in `.mise/tasks/`. The filename is the task name. Subdirectories create namespaced tasks (e.g., `.mise/tasks/build/backend` → `mise run build:backend`).

### Task structure

```bash
#!/usr/bin/env bash
set -euo pipefail
#MISE description="What this task does"
#MISE depends=["other-task"]

# task body
```

Every task script MUST:
1. Start with `#!/usr/bin/env bash`
2. Set `set -euo pipefail` (fail fast, no unset vars, pipe failures)
3. Include `#MISE description="..."` for discoverability
4. Be executable (`chmod +x`)

### Task metadata comments

```bash
#MISE description="Build the Go backend"
#MISE alias="b"
#MISE depends=["lint"]
#MISE sources=["backend/**/*.go", "backend/go.mod"]
#MISE outputs=["bin/cubby"]
#MISE env={CGO_ENABLED="0"}
#MISE dir="backend"
```

| Directive    | Purpose                                              |
|-------------|------------------------------------------------------|
| `description` | Shown in `mise tasks` listing                      |
| `alias`      | Short name (`mise run b`)                           |
| `depends`    | Run these tasks first                               |
| `wait_for`   | Wait for these if running, but don't trigger them   |
| `sources`    | Input files — skip task if outputs are newer        |
| `outputs`    | Output files — used with `sources` for caching      |
| `env`        | Environment variables for this task                 |
| `dir`        | Working directory (relative to project root)        |
| `tools`      | Required tool versions                              |

### Task grouping with directories

```
.mise/tasks/
├── build               # mise run build
├── build/
│   ├── backend         # mise run build:backend
│   ├── frontend        # mise run build:frontend
│   └── pi              # mise run build:pi
├── dev                 # mise run dev
├── dev/
│   ├── backend         # mise run dev:backend
│   └── frontend        # mise run dev:frontend
├── test                # mise run test
├── lint                # mise run lint
├── fmt                 # mise run fmt
├── docker/
│   ├── build           # mise run docker:build
│   ├── up              # mise run docker:up
│   └── down            # mise run docker:down
├── flyme/
│   ├── up              # mise run flyme:up
│   ├── down            # mise run flyme:down
│   ├── status          # mise run flyme:status
│   ├── check           # mise run flyme:check
│   ├── redo            # mise run flyme:redo
│   └── create          # mise run flyme:create -- "<description>"
├── db/
│   ├── backup          # mise run db:backup [output-path]
│   └── restore         # mise run db:restore -- --yes <backup-file>
└── clean               # mise run clean
```

**Note**: A file and directory can share a name. `.mise/tasks/build` (the script) runs `mise run build`, while `.mise/tasks/build/backend` runs `mise run build:backend`.

### Running tasks

```sh
mise run build              # run a task
mise run build:backend      # run a namespaced task
mise run lint test          # run multiple tasks
mise run test -- -v         # pass args through to the script
mise tasks                  # list all available tasks
```

Arguments after `--` are passed as `$@` to the script.

## Environment variables

```toml
[env]
CUBBY_PORT = "8080"
CUBBY_DB_PATH = "./data/cubby.db"

# Load from .env file
[env]
_.file = ".env"
```

## mise.toml configuration

```toml
# mise.toml — project root
[tools]
go = "1.25"
node = "22"

[env]
CGO_ENABLED = "0"

[task_config]
dir = "{{cwd}}"  # default task working directory
```

## Rules

- **All tools via mise** — never install dev tools globally or via Homebrew for project use
- **Tasks over Makefile** — use `.mise/tasks/` scripts, not Make targets
- **Pin languages, float tools** — `go = "1.25"` but `gofumpt = "latest"`
- **`set -euo pipefail`** — every task script, no exceptions
- **Descriptions on every task** — `#MISE description="..."` for `mise tasks` output

## Don'ts

- Don't use `[tasks]` in `mise.toml` for complex logic — use file-based tasks
- Don't skip `set -euo pipefail` — silent failures are bugs
- Don't hardcode tool paths — mise shims handle this
- Don't mix Makefile and mise tasks — pick one (mise)
