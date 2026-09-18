# Cubby

Cubby is a hobby project for managing home inventory with a Go backend, SQLite storage, and a modern frontend.

## Project status

- Personal/open-source side project
- Active development
- Breaking changes may happen while features are evolving

## What Cubby aims to do

- Track items, locations, and tags
- Support search across stored inventory
- Generate QR codes for quick lookup workflows

## Tech stack

- Go 1.25 backend (`net/http`)
- SQLite (`modernc.org/sqlite`)
- Frontend app (Node 22 / npm)
- Task runner: `mise`

## Getting started

### Prerequisites

- [mise](https://mise.jdx.dev/)
- Go 1.25+
- Node.js 22+
- `sqlite3` CLI (for database backup and restore)

### Run locally

```sh
mise run dev
```

Useful commands:

```sh
mise run test
mise run build
mise run verify

# Flyme migrations
mise run flyme:up
mise run flyme:status
mise run flyme:check  # read-only migration file validation
mise run flyme:down -- 1
mise run flyme:redo
mise run flyme:create -- "add item barcode"  # create new migration files

# Database backup / restore
mise run db:backup                            # backup to default timestamped path
mise run db:backup ./backups/pre-deploy.db    # backup to specific path
mise run db:restore -- --yes --stopped ./backups/pre-deploy.db  # restore (app must be stopped)
```

## Repository layout

```text
backend/
  cmd/server/          # API entry point
  internal/api/        # HTTP handlers
  internal/service/    # Business logic
  internal/db/         # SQLite queries + migrations
  internal/models/     # Shared domain models
frontend/              # Frontend app
```

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before opening a PR.

## Security

If you find a vulnerability, please follow [SECURITY.md](./SECURITY.md).

## Code of conduct

This project follows the [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).
