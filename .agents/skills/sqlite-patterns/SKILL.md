---
name: sqlite-patterns
description: SQLite database patterns for Cubby — migrations, queries, transactions, FTS5 search, dynamic updates
---

## When to use

Use when adding or modifying database code in `backend/internal/db/`.

## Connection

```go
db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
```

Always enable: WAL mode, busy timeout, foreign keys.

## Migrations

Embedded SQL files in `internal/db/migrations/`. Naming: `NNN_description.up.sql` / `.down.sql`.

```sql
-- 002_add_labels.up.sql
CREATE TABLE IF NOT EXISTS labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#6b7280',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Register new migrations in `db.go`:

```go
migrations := []string{"001_initial", "002_add_labels"}
```

## Query methods

All queries live on the `Queries` struct. Follow the verb convention:

| Method | Signature |
|--------|-----------|
| List   | `func (q *Queries) ListThings() ([]models.Thing, error)` |
| Get    | `func (q *Queries) GetThing(id int64) (*models.Thing, error)` |
| Create | `func (q *Queries) CreateThing(req models.CreateThingRequest) (*models.Thing, error)` |
| Update | `func (q *Queries) UpdateThing(id int64, req models.UpdateThingRequest) (*models.Thing, error)` |
| Delete | `func (q *Queries) DeleteThing(id int64) error` |

## Not found

Return `nil, nil` — not found is not an error:

```go
if err == sql.ErrNoRows {
    return nil, nil
}
```

## Transactions

```go
tx, err := q.db.Begin()
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback()

// ... operations on tx ...

return tx.Commit()
```

## Dynamic partial updates

```go
sets := []string{}
args := []any{}

if req.Name != nil {
    sets = append(sets, "name = ?")
    args = append(args, *req.Name)
}

if len(sets) == 0 {
    return q.GetThing(id) // no changes
}

sets = append(sets, "updated_at = ?")
args = append(args, time.Now().UTC())
args = append(args, id)

query := fmt.Sprintf("UPDATE things SET %s WHERE id = ?", strings.Join(sets, ", "))
```

## FTS5 search

When adding a searchable entity, create an FTS5 virtual table with triggers:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS things_fts USING fts5(
    name, description,
    content='things',
    content_rowid='id'
);

-- Keep FTS in sync
CREATE TRIGGER things_ai AFTER INSERT ON things BEGIN
    INSERT INTO things_fts(rowid, name, description) VALUES (new.id, new.name, new.description);
END;

CREATE TRIGGER things_ad AFTER DELETE ON things BEGIN
    INSERT INTO things_fts(things_fts, rowid, name, description) VALUES ('delete', old.id, old.name, old.description);
END;

CREATE TRIGGER things_au AFTER UPDATE ON things BEGIN
    INSERT INTO things_fts(things_fts, rowid, name, description) VALUES ('delete', old.id, old.name, old.description);
    INSERT INTO things_fts(rowid, name, description) VALUES (new.id, new.name, new.description);
END;
```

## Row scanning

Always `defer rows.Close()` immediately. Check `rows.Err()` after the loop:

```go
rows, err := q.db.Query(query, args...)
if err != nil {
    return nil, fmt.Errorf("list things: %w", err)
}
defer rows.Close()

var things []models.Thing
for rows.Next() {
    var t models.Thing
    if err := rows.Scan(&t.ID, &t.Name, ...); err != nil {
        return nil, fmt.Errorf("scan thing: %w", err)
    }
    things = append(things, t)
}
return things, rows.Err()
```

## Don'ts

- No ORM — raw SQL only
- No `*` in SELECT — always list columns explicitly
- No `interface{}` for args — use `[]any`
- Never ignore `rows.Err()`
