---
name: flyme
description: Safe, repeatable SQLite data migration workflow for Cubby across development and production
---

## Overview

Use when planning, writing, rehearsing, and shipping schema/data migrations in Cubby.

Scope:
- Go backend (`backend/`)
- SQLite database (`modernc.org/sqlite`)
- Task runner (`mise`)

Goals:
- Forward-only, auditable migration history
- Repeatable deploys across dev/staging/prod
- Production-safe rollout with explicit verification and recovery paths

## Migration Strategy

### Principles

1. **Forward-only first**
   - Prefer additive changes and follow-up cleanup migrations.
   - Avoid relying on down migrations in production.

2. **Idempotent and auditable**
   - Re-running migration/recovery scripts should be safe.
   - Every production SQL change must be committed in versioned files.

3. **Separate schema and data changes when possible**
   - Migration A: schema expand (new columns/tables/indexes).
   - Migration B: backfill in controlled batches.
   - Migration C: enforce stricter constraints after backfill.

4. **Zero-downtime preference**
   - Use expand → migrate data → contract pattern.
   - Keep old and new app versions compatible during transition.

### Safety guardrails

- No destructive migration without confirmed backup + restore drill.
- No manual hotfix SQL in production without recording it in versioned SQL.
- Do not mark deploy complete until verification gates pass.

## Dev Workflow

```sh
mise run flyme:create -- "add item barcode"  # create new migration files
```

This generates `NNN_add_item_barcode.up.sql` and `NNN_add_item_barcode.down.sql` in
`backend/internal/db/migrations/` with the next sequential number.

Then apply locally:

```sh
mise run flyme:up
```

Then run full validation:

```sh
mise run verify
```

### 3) Seed/backfill safely in dev

- Use deterministic test fixtures.
- Make backfill scripts restartable (checkpoint by primary key range).
- Commit backfill SQL/script in repo (no one-off local SQL only).

Example backfill pattern:

```sql
-- Repeat in batches: WHERE id > ? ORDER BY id LIMIT 1000
UPDATE items
SET barcode = printf('ITEM-%06d', id)
WHERE barcode IS NULL
  AND id BETWEEN ? AND ?;
```

### 4) Validate app compatibility

Run:

```sh
mise run test
mise run build
```

Ensure both read/write paths work with pre- and post-backfill states.

### 5) Rollback simulation (dev only)

- Practice restoration from backup copy.
- Verify app starts and tests pass after restore.
- Prefer forward-fix strategy validation even if down SQL exists.

## Prod Workflow

### 1) Preflight checks

- Backup exists and restore tested recently.
  ```sh
  mise run db:backup ./backups/pre-deploy.db           # take a fresh backup
  mise run db:restore -- --yes --stopped ./backups/pre-deploy.db  # verify restore works (app must be stopped)
  ```
- Sufficient disk space for DB + WAL growth during migration.
- No long-running lock-heavy jobs.
- Deployed app version is compatible with migration stage.

### 2) Dry-run / staging rehearsal

- Rehearse against staging snapshot before production.
- Record duration, lock behavior, row counts, and checkpoints.
- Define abort threshold (time, lock contention, error rate).

### 3) Deployment order

Default order for zero-downtime:
1. Deploy backward-compatible app code.
2. Apply additive schema migration.
3. Run online backfill in batches.
4. Enable stricter constraints/cleanup in later release.

Why: avoids breaking old readers/writers during rollout.

### 4) Online backfill strategy (large tables)

- Process bounded batches (`LIMIT`, key ranges).
- Store checkpoint (`last_processed_id`) after each batch.
- Retry transient failures with bounded retries.
- Pause between batches when lock pressure rises.

### 5) Post-migration verification

- Run schema/version check.
- Run row-count and null-check validation queries.
- Run app health checks and key endpoint smoke tests.

## Rollback & Recovery

### Choose rollback vs forward-fix

- **Safe rollback**: migration was additive and no irreversible writes consumed by new app behavior.
- **Forward-fix required**: destructive/irreversible data change or mixed-version writes already happened.

### Restore-from-backup procedure

1. Put app into maintenance/read-only mode.
2. Stop writers.
3. Restore DB backup:
   ```sh
   mise run db:restore -- --yes --stopped <backup-file>  # app must be stopped
   ```
4. Restart app on known-compatible release.
5. Run verification checklist before reopening writes.

### Incident checklist

- Timestamp and impact window captured.
- Migration version + git SHA recorded.
- Failing query/error logs preserved.
- Recovery decision (rollback vs forward-fix) documented.
- Post-incident follow-up migration/task created.

## Verification Checklist

Use this as Definition of Done for migration rollout:

- [ ] Backup verified
- [ ] Migration applied
- [ ] Backfill complete
- [ ] App healthy
- [ ] Verification queries passed
- [ ] Rollback plan validated

Recommended verification commands:

```sh
mise run flyme:status
mise run flyme:check
mise run verify
mise run build
mise run test
```

## Troubleshooting

### Locked database

- **Symptom:** migration/backfill stalls with lock/busy errors.
- **Cause:** competing write transactions or long-running readers.
- **Fix:** pause writers, reduce batch size, retry with checkpoints, rerun during lower traffic.

### Long-running migration

- **Symptom:** migration exceeds expected window.
- **Cause:** large table rewrite, missing index support, oversized batch.
- **Fix:** split into expand + batched backfill, add supporting indexes in prior migration, throttle batch size.

### Partial backfill

- **Symptom:** some rows remain unmigrated (`NULL`/default values).
- **Cause:** interrupted job, non-restartable script, missing checkpoint logic.
- **Fix:** resume from persisted checkpoint, rerun idempotent batch job, validate with gap queries.

### Failed deploy mismatch (app vs schema)

- **Symptom:** app errors after deploy due to missing/renamed columns.
- **Cause:** non-compatible deploy order or premature contract migration.
- **Fix:** redeploy compatible app version, apply additive schema first, defer destructive contract step to later release.
