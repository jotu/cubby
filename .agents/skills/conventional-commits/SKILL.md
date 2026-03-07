---
name: conventional-commits
description: Conventional Commits v1.0.0 — structured commit messages for humans and machines
---

## Spec

<https://www.conventionalcommits.org/en/v1.0.0/>

## Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

- **Description**: imperative mood, lowercase, no period, max 72 chars
- **Body**: blank line after description. Free-form, wrap at 80 chars. Explain *why* not *what*. May contain multiple paragraphs.
- **Footer(s)**: blank line after body. Format: `Token: value` or `Token #value`. Use `-` in place of spaces in tokens (e.g., `Reviewed-by`, `Acked-by`).

## Types

| Type       | When                                          | SemVer   |
|------------|-----------------------------------------------|----------|
| `feat`     | New feature or capability                     | MINOR    |
| `fix`      | Bug fix                                       | PATCH    |
| `refactor` | Code change that neither fixes nor adds       | —        |
| `test`     | Adding or updating tests                      | —        |
| `docs`     | Documentation only                            | —        |
| `build`    | Build system, dependencies                    | —        |
| `ci`       | CI/CD configuration                           | —        |
| `chore`    | Other maintenance (tooling, config)           | —        |
| `style`    | Formatting, whitespace (no logic change)      | —        |
| `perf`     | Performance improvement                       | PATCH    |
| `revert`   | Revert a previous commit                      | varies   |

Any type with a `BREAKING CHANGE` → MAJOR regardless of type.

## Scopes (this project)

| Scope      | Area                                          |
|------------|-----------------------------------------------|
| `db`       | `internal/db/` — queries, migrations          |
| `api`      | `internal/api/` — HTTP handlers               |
| `models`   | `internal/models/` — types, request structs   |
| `service`  | `internal/service/` — business logic          |
| `qr`       | `internal/qr/` — QR code generation           |
| `frontend` | `frontend/` — UI                              |
| `docker`   | Dockerfile, docker-compose                    |

Omit scope when change spans multiple areas.

## Breaking changes

Two ways to indicate (both valid, can be combined):

```
# 1. ! marker in subject line
feat(api)!: change auth token format

# 2. BREAKING CHANGE footer (MUST be uppercase)
feat(api): change auth token format

BREAKING CHANGE: tokens issued before v2 are no longer accepted
```

`BREAKING-CHANGE` (with hyphen) is synonymous with `BREAKING CHANGE`.

When using `!`, the `BREAKING CHANGE:` footer MAY be omitted — the description serves as the breaking change explanation.

## Footers

Footers follow [git trailer format](https://git-scm.com/docs/git-interpret-trailers):

```
Token: value
Token #value
```

Common footers:

| Footer             | Use                                    |
|--------------------|----------------------------------------|
| `BREAKING CHANGE:` | Breaking API change (MUST be uppercase)|
| `Refs: #123`       | Related issue                          |
| `Closes #456`      | Closes issue on merge                  |
| `Reviewed-by: X`   | Reviewer attribution                   |

## Rules

- **One concern per commit** — never mix a feature with a refactor
- **Atomic** — if a commit touches two unrelated things, split it
- **Subject line**: lowercase type, imperative mood, no period, ≤72 chars
- **Body**: wrap at 80 chars, explain *why* not *what*
- **Type required** — every commit MUST have a type prefix

## Examples

### Feature with body

```
feat(db): add label table and queries

Adds labels entity with CRUD operations and FTS5 search.
Migration 002_add_labels creates the table and triggers.
```

### Bug fix

```
fix(api): return 404 for missing items

GetItem returned 500 when item didn't exist because nil check
was missing after the query.
```

### Refactor (no body needed for small changes)

```
refactor(db): extract common scan helper
```

### Breaking change with footer

```
feat(api)!: require auth token for all endpoints

All endpoints now require a valid Bearer token. Anonymous
access is no longer supported.

BREAKING CHANGE: unauthenticated requests return 401 instead of proceeding
Refs: #42
```

### Revert

```
revert: add label table and queries

Refs: a215868
```

### Build / dependency change

```
build: upgrade Go to 1.25
```

## PR titles

PR title = squash-merge commit message. Same format:

```
feat(api): add item CRUD endpoints
fix(db): handle concurrent writes with busy timeout
build: upgrade Go to 1.25
docs: add API usage examples to README
```

## PR body

```markdown
## Summary
- <1-3 bullet points describing what and why>

## Test plan
- <how this was verified>
```

## Don'ts

- Don't capitalize the description (`Fix bug` → `fix bug`)
- Don't end the subject with a period
- Don't use past tense (`added` → `add`, `fixed` → `fix`)
- Don't mix concerns in one commit
- Don't use `chore` as a catch-all — prefer `build`, `ci`, `docs` when they fit
