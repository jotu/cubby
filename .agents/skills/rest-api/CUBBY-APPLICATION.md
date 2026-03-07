# REST API Design Application Guide for Cubby

This guide shows how to apply the REST API best practices to Cubby's specific domain (home inventory system).

## API Structure

Based on the AGENTS.md domain:

### Core Resources

**Locations** (home organization structure)
- Plural: `/locations`
- Hierarchy: `/locations/{id}` (parent ID in body for nesting)
- Examples: Kitchen, Bedroom, Storage Room

**Items** (inventory objects)
- Plural: `/items`
- Child of locations: `/items?location_id={id}`
- References location but can list all items too

**QR Codes** (for scanning)
- Plural: `/qr-codes`
- Linked to items: `POST /items/{id}/qr-code`

---

## API Endpoint Examples

### Locations API

#### List Locations
```
GET /locations?limit=20&cursor=abc123

Response: 200 OK
{
  "items": [
    {
      "id": 1,
      "name": "Kitchen",
      "description": "Main kitchen",
      "parentId": null,
      "createdAt": "2026-03-05T10:30:00Z",
      "updatedAt": "2026-03-05T10:30:00Z"
    }
  ],
  "pagination": {
    "self": "https://api.cubby.local/locations?limit=20&cursor=abc123",
    "next": "https://api.cubby.local/locations?limit=20&cursor=def456",
    "has_more": true
  }
}
```

#### Get Single Location
```
GET /locations/1

Response: 200 OK
{
  "id": 1,
  "name": "Kitchen",
  "description": "Main kitchen",
  "parentId": null,
  "createdAt": "2026-03-05T10:30:00Z",
  "updatedAt": "2026-03-05T10:30:00Z"
}
```

#### Create Location
```
POST /locations
Content-Type: application/json
{
  "name": "Pantry",
  "description": "Kitchen pantry",
  "parentId": 1
}

Response: 201 Created
Location: /locations/2
{
  "id": 2,
  "name": "Pantry",
  "description": "Kitchen pantry",
  "parentId": 1,
  "createdAt": "2026-03-05T10:35:00Z",
  "updatedAt": "2026-03-05T10:35:00Z"
}
```

#### Update Location (Partial)
```
PATCH /locations/2
Content-Type: application/merge-patch+json
{
  "description": "Updated pantry description"
}

Response: 200 OK
{
  "id": 2,
  "name": "Pantry",
  "description": "Updated pantry description",
  "parentId": 1,
  "createdAt": "2026-03-05T10:35:00Z",
  "updatedAt": "2026-03-05T10:38:00Z"
}
```

#### Delete Location
```
DELETE /locations/2

Response: 204 No Content
```

### Items API

#### List Items (with filtering & pagination)
```
GET /items?location_id=1&category=kitchen&limit=20&cursor=xyz

Response: 200 OK
{
  "items": [
    {
      "id": 1,
      "name": "Laptop",
      "category": "electronics",
      "locationId": 1,
      "quantity": 1,
      "description": "Work laptop",
      "createdAt": "2026-03-05T10:30:00Z",
      "updatedAt": "2026-03-05T10:30:00Z"
    }
  ],
  "pagination": {
    "self": "...",
    "next": "...",
    "has_more": false
  }
}
```

#### Create Item (with Idempotency-Key)
```
POST /items
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440001
Content-Type: application/json
{
  "name": "Monitor",
  "category": "electronics",
  "locationId": 1,
  "quantity": 2,
  "description": "4K Monitor"
}

Response: 201 Created
Location: /items/2
{
  "id": 2,
  "name": "Monitor",
  "category": "electronics",
  "locationId": 1,
  "quantity": 2,
  "description": "4K Monitor",
  "createdAt": "2026-03-05T10:35:00Z",
  "updatedAt": "2026-03-05T10:35:00Z"
}

# Retry with same Idempotency-Key
POST /items
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440001
Content-Type: application/json
{ "name": "Monitor", ... }

Response: 200 OK  # Returns same as first, not 201
{
  "id": 2,
  ...
}
```

#### Search Items
```
GET /items?q=laptop&sort=-createdAt

Response: 200 OK
{
  "items": [...]
}
```

#### Bulk Create Items
```
POST /items/batch
Content-Type: application/json
{
  "items": [
    { "name": "Item 1", "category": "kitchen", "locationId": 1, "quantity": 1 },
    { "name": "Item 2", "category": "bedroom", "locationId": 2, "quantity": 3 }
  ]
}

Response: 207 Multi-Status
{
  "items": [
    { "index": 0, "status": 201, "data": { "id": 3, "name": "Item 1", ... } },
    { "index": 1, "status": 201, "data": { "id": 4, "name": "Item 2", ... } }
  ]
}
```

### QR Code API

#### Generate QR Code for Item
```
POST /items/1/qr-code

Response: 202 Accepted
Location: /qr-codes/123
{
  "id": "123",
  "itemId": 1,
  "status": "generating",
  "createdAt": "2026-03-05T10:40:00Z"
}

# Poll for completion
GET /qr-codes/123

Response: 200 OK
{
  "id": "123",
  "itemId": 1,
  "status": "ready",
  "data": "iVBORw0KGgoAAAANS...",  // Base64 PNG
  "url": "/items/1/qr-code.png",
  "createdAt": "2026-03-05T10:40:00Z"
}
```

---

## Error Response Examples

### Validation Error (400 Bad Request)
```
POST /items
{ "name": "", "categoryId": "invalid" }

Response: 400 Bad Request
Content-Type: application/problem+json
{
  "type": "https://api.cubby.local/probs/validation-error",
  "title": "Validation Failed",
  "status": 400,
  "detail": "Required fields are missing or invalid",
  "instance": "/items",
  "timestamp": "2026-03-05T10:40:00Z",
  "errors": [
    {
      "field": "name",
      "message": "Name is required and must not be empty"
    },
    {
      "field": "category",
      "message": "Category must be one of: electronics, kitchen, bedroom, storage"
    }
  ]
}
```

### Not Found (404)
```
GET /items/999999

Response: 404 Not Found
Content-Type: application/problem+json
{
  "type": "https://api.cubby.local/probs/not-found",
  "title": "Item Not Found",
  "status": 404,
  "detail": "Item with ID 999999 does not exist",
  "instance": "/items/999999",
  "timestamp": "2026-03-05T10:40:00Z"
}
```

### Rate Limit (429)
```
GET /items

Response: 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1614876300
Retry-After: 60
Content-Type: application/problem+json
{
  "type": "https://api.cubby.local/probs/rate-limited",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "You have exceeded 100 requests per hour",
  "instance": "/items",
  "timestamp": "2026-03-05T10:40:00Z"
}
```

### Conflict (409) - Optimistic Locking
```
PATCH /items/1
If-Match: "v1.0.0"
{ "quantity": 5 }

Response: 409 Conflict
Content-Type: application/problem+json
{
  "type": "https://api.cubby.local/probs/conflict",
  "title": "Version Conflict",
  "status": 409,
  "detail": "Item was modified since your last read. Fetch it again and retry.",
  "instance": "/items/1",
  "timestamp": "2026-03-05T10:40:00Z"
}
```

---

## Go Implementation Patterns

### Location Handler (Example)

```go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/joacim/cubby/internal/models"
	"github.com/joacim/cubby/internal/service"
)

type LocationHandler struct {
	svc *service.LocationService
}

// GET /locations
func (h *LocationHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 20, 100)
	cursor := r.URL.Query().Get("cursor")

	locations, nextCursor, err := h.svc.ListLocations(r.Context(), cursor, limit)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError,
			"https://api.cubby.local/probs/server-error",
			"Server Error",
			fmt.Sprintf("failed to list locations: %v", err),
			r.RequestURI,
		)
		return
	}

	type Response struct {
		Items      []*models.Location `json:"items"`
		Pagination struct {
			Self    string `json:"self"`
			Next    string `json:"next,omitempty"`
			HasMore bool   `json:"has_more"`
		} `json:"pagination"`
	}

	resp := Response{
		Items: locations,
	}
	resp.Pagination.Self = r.RequestURI
	if nextCursor != "" {
		resp.Pagination.Next = fmt.Sprintf("/locations?limit=%d&cursor=%s", limit, nextCursor)
		resp.Pagination.HasMore = true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// POST /locations
func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest,
			"https://api.cubby.local/probs/validation-error",
			"Validation Failed",
			"Invalid JSON in request body",
			r.RequestURI,
		)
		return
	}

	// Validate
	if err := validateLocation(&req); err != nil {
		writeProblem(w, http.StatusBadRequest,
			"https://api.cubby.local/probs/validation-error",
			"Validation Failed",
			err.Error(),
			r.RequestURI,
		)
		return
	}

	loc, err := h.svc.CreateLocation(r.Context(), &req)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError,
			"https://api.cubby.local/probs/server-error",
			"Server Error",
			"Failed to create location",
			r.RequestURI,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/locations/%d", loc.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loc)
}

// PATCH /locations/{id}
func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.PathValue("id"))
	if id == 0 {
		writeProblem(w, http.StatusBadRequest,
			"https://api.cubby.local/probs/validation-error",
			"Invalid ID",
			"Location ID must be a positive integer",
			r.RequestURI,
		)
		return
	}

	var req models.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest,
			"https://api.cubby.local/probs/validation-error",
			"Validation Failed",
			"Invalid JSON in request body",
			r.RequestURI,
		)
		return
	}

	loc, err := h.svc.UpdateLocation(r.Context(), id, &req)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			writeProblem(w, http.StatusNotFound,
				"https://api.cubby.local/probs/not-found",
				"Location Not Found",
				fmt.Sprintf("Location %d does not exist", id),
				r.RequestURI,
			)
		default:
			writeProblem(w, http.StatusInternalServerError,
				"https://api.cubby.local/probs/server-error",
				"Server Error",
				"Failed to update location",
				r.RequestURI,
			)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loc)
}

// DELETE /locations/{id}
func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.PathValue("id"))
	if id == 0 {
		writeProblem(w, http.StatusBadRequest,
			"https://api.cubby.local/probs/validation-error",
			"Invalid ID",
			"Location ID must be a positive integer",
			r.RequestURI,
		)
		return
	}

	if err := h.svc.DeleteLocation(r.Context(), id); err != nil {
		switch {
		case err == sql.ErrNoRows:
			writeProblem(w, http.StatusNotFound,
				"https://api.cubby.local/probs/not-found",
				"Location Not Found",
				fmt.Sprintf("Location %d does not exist", id),
				r.RequestURI,
			)
		default:
			writeProblem(w, http.StatusInternalServerError,
				"https://api.cubby.local/probs/server-error",
				"Server Error",
				"Failed to delete location",
				r.RequestURI,
			)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

### Problem Details Helper

```go
type ProblemDetail struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Status    int                    `json:"status"`
	Detail    string                 `json:"detail"`
	Instance  string                 `json:"instance"`
	Timestamp time.Time              `json:"timestamp"`
	Extra     map[string]interface{} `json:"-"`
}

func (p *ProblemDetail) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":      p.Type,
		"title":     p.Title,
		"status":    p.Status,
		"detail":    p.Detail,
		"instance":  p.Instance,
		"timestamp": p.Timestamp.Format(time.RFC3339),
	}
	for k, v := range p.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

func writeProblem(w http.ResponseWriter, status int, typ, title, detail, instance string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ProblemDetail{
		Type:      typ,
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now().UTC(),
	})
}
```

---

## Implementation Checklist for Cubby Endpoints

When building each endpoint:

- [ ] **Resource naming**: Plural, kebab-case (e.g., `/locations`, `/qr-codes`)
- [ ] **HTTP method**: GET (read), POST (create), PATCH (update), DELETE (delete)
- [ ] **Status codes**: 200/201/204 for success, 400/404/409 for errors
- [ ] **Errors**: RFC 7807 Problem Details format
- [ ] **Pagination**: Cursor-based for collection endpoints
- [ ] **Idempotency**: `Idempotency-Key` header for POST endpoints
- [ ] **Response format**: Flat JSON, no unnecessary wrapper
- [ ] **Headers**: `Content-Type: application/json`, `Location` for 201
- [ ] **Tests**: Happy path + error cases (validation, not found, conflicts)
- [ ] **Documentation**: Document in OpenAPI/Swagger

---

## Current Status

✅ REST API skill guide created
✅ Examples specific to Cubby domain
✅ Go implementation patterns provided
⏳ Ready for use in new endpoint development

## Next Steps

1. Start implementing endpoints using these patterns
2. Create OpenAPI specification for Cubby API
3. Add more endpoints following this pattern
4. Document in project README
