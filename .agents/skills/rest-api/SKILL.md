---
name: rest-api
description: REST API conventions for Cubby
---

## When to use

Use when designing or reviewing HTTP endpoints for locations, items, and tags.

## Resource naming

- Plural nouns: `/locations`, `/items`, `/tags`
- Kebab-case URLs: `/item-tags`
- camelCase JSON: `{ "locationId": 123 }`

## HTTP methods

| Method | Semantics | Status |
| --- | --- | --- |
| GET | Read resource(s) | 200 |
| POST | Create in collection | 201 |
| PUT | Replace entire resource | 200/204 |
| PATCH | Partial update | 200/204 |
| DELETE | Remove resource | 204 |

## Status codes quick reference

| Code | Use |
| --- | --- |
| 200 | Success with body |
| 201 | Created with Location |
| 204 | Success, no body |
| 400 | Invalid request payload/params |
| 404 | Resource not found |
| 409 | Conflict (version/unique) |
| 422 | Semantic validation failure |
| 500 | Unexpected server error |

## Error response format

```json
{
  "code": "validation_error",
  "message": "Name is required",
  "details": { "field": "name" }
}
```

## Pagination (cursor-based)

Query params: `cursor`, `limit`.

```json
{
  "items": [ { "id": 1, "name": "Laptop" } ],
  "pagination": {
    "cursor": "next-token",
    "limit": 20,
    "hasMore": true
  }
}
```

## Filtering and sorting

- Filters: `?locationId=123&tagId=7`
- Ranges: `?createdAfter=2024-01-01&createdBefore=2024-12-31`
- Sorting: `?sort=name` or `?sort=-createdAt,name`

## Partial updates with PATCH

Use merge semantics; only provided fields change.
Missing fields stay unchanged; explicit `null` clears a field.

## Nested vs flat resources

Nest when child is scoped to parent:
`/locations/123/items`

Flatten when item can stand alone:
`/items?locationId=123`

## Versioning

Use `/v1` in the URL for breaking changes:
`/v1/locations`, `/v1/items`, `/v1/tags`.

## Rules

- Keep URLs noun-based and kebab-case.
- Use camelCase for JSON fields.
- Return 201 with `Location` on create.
- Always support pagination on list endpoints.
- Prefer PATCH for partial updates.

## Don'ts

- Don’t use verbs in URLs (`/getItems`).
- Don’t return 200 for deletes with no body (use 204).
- Don’t mix nested and flat styles for the same relationship.
