package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/joacim/cubby/internal/db"
	"github.com/joacim/cubby/internal/models"
)

func setupTestServer(t *testing.T) (*Server, *db.Queries) {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	uploadDir := t.TempDir()
	queries := db.NewQueries(database)
	server := NewServer(queries, uploadDir)
	return server, queries
}

func createLocation(t *testing.T, queries *db.Queries, name string) *models.Location {
	t.Helper()

	loc, err := queries.CreateLocation(models.CreateLocationRequest{Name: name})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return loc
}

func createItem(t *testing.T, queries *db.Queries, name string, locationID *int64) *models.Item {
	t.Helper()

	item, err := queries.CreateItem(models.CreateItemRequest{
		Name:       name,
		LocationID: locationID,
		Quantity:   1,
		Tags:       []string{},
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	return item
}

func TestHandleListLocations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		createLocation(t, queries, "Garage")
		createLocation(t, queries, "Basement")

		req := httptest.NewRequest(http.MethodGet, "/v1/locations", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var locations []models.Location
		if err := json.NewDecoder(rec.Body).Decode(&locations); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(locations) != 2 {
			t.Errorf("locations len = %d, want %d", len(locations), 2)
		}
	})
}

func TestHandleCreateLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, _ := setupTestServer(t)

		body := strings.NewReader(`{"name":"Garage"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/locations", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		if rec.Header().Get("Location") == "" {
			t.Errorf("missing Location header")
		}

		var loc models.Location
		if err := json.NewDecoder(rec.Body).Decode(&loc); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if loc.Name != "Garage" {
			t.Errorf("name = %q, want %q", loc.Name, "Garage")
		}
	})
}

func TestHandleCreateLocation_InvalidJSON(t *testing.T) {
	t.Run("invalid_json", func(t *testing.T) {
		server, _ := setupTestServer(t)

		body := strings.NewReader("{")
		req := httptest.NewRequest(http.MethodPost, "/v1/locations", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleCreateLocation_MissingName(t *testing.T) {
	t.Run("missing_name", func(t *testing.T) {
		server, _ := setupTestServer(t)

		body := strings.NewReader(`{"name":""}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/locations", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleGetLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		loc := createLocation(t, queries, "Garage")

		req := httptest.NewRequest(http.MethodGet, "/v1/locations/"+strconv.FormatInt(loc.ID, 10), nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var got models.Location
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.ID != loc.ID {
			t.Errorf("id = %d, want %d", got.ID, loc.ID)
		}
	})
}

func TestHandleGetLocation_NotFound(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		server, _ := setupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/v1/locations/9999", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestHandleUpdateLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		loc := createLocation(t, queries, "Garage")

		body := strings.NewReader(`{"name":"Shed"}`)
		req := httptest.NewRequest(http.MethodPatch, "/v1/locations/"+strconv.FormatInt(loc.ID, 10), body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var got models.Location
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.Name != "Shed" {
			t.Errorf("name = %q, want %q", got.Name, "Shed")
		}
	})
}

func TestHandleDeleteLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		loc := createLocation(t, queries, "Garage")

		req := httptest.NewRequest(http.MethodDelete, "/v1/locations/"+strconv.FormatInt(loc.ID, 10), nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func TestHandleListItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		createItem(t, queries, "Hammer", nil)
		createItem(t, queries, "Wrench", nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/items", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var items []models.Item
		if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("items len = %d, want %d", len(items), 2)
		}
	})
}

func TestHandleCreateItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)
		loc := createLocation(t, queries, "Garage")

		body := strings.NewReader(`{"name":"Hammer","locationId":` + strconv.FormatInt(loc.ID, 10) + `,"quantity":2,"tags":[]}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/items", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		if rec.Header().Get("Location") == "" {
			t.Errorf("missing Location header")
		}

		var item models.Item
		if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if item.Name != "Hammer" {
			t.Errorf("name = %q, want %q", item.Name, "Hammer")
		}
	})
}

func TestHandleGetItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		item := createItem(t, queries, "Hammer", nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/items/"+strconv.FormatInt(item.ID, 10), nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var got models.Item
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.ID != item.ID {
			t.Errorf("id = %d, want %d", got.ID, item.ID)
		}
	})
}

func TestHandleGetItem_NotFound(t *testing.T) {
	t.Run("not_found", func(t *testing.T) {
		server, _ := setupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/v1/items/9999", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestHandleUpdateItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		item := createItem(t, queries, "Hammer", nil)

		body := strings.NewReader(`{"name":"Mallet"}`)
		req := httptest.NewRequest(http.MethodPatch, "/v1/items/"+strconv.FormatInt(item.ID, 10), body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var got models.Item
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.Name != "Mallet" {
			t.Errorf("name = %q, want %q", got.Name, "Mallet")
		}
	})
}

func TestHandleDeleteItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		item := createItem(t, queries, "Hammer", nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/items/"+strconv.FormatInt(item.ID, 10), nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func TestHandleUploadPhoto(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		item := createItem(t, queries, "Hammer", nil)

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		fw, err := writer.CreateFormFile("photo", "test.jpg")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := fw.Write([]byte("fake image data")); err != nil {
			t.Fatalf("write form file: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/items/"+strconv.FormatInt(item.ID, 10)+"/photo", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp struct {
			PhotoPath string `json:"photoPath"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.PhotoPath == "" {
			t.Errorf("photoPath is empty")
		}

		getReq := httptest.NewRequest(http.MethodGet, "/v1/items/"+strconv.FormatInt(item.ID, 10), nil)
		getRec := httptest.NewRecorder()
		server.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusOK {
			t.Fatalf("get item status = %d, want %d", getRec.Code, http.StatusOK)
		}

		var got models.Item
		if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
			t.Fatalf("decode get response: %v", err)
		}
		if got.PhotoPath == "" {
			t.Errorf("photoPath not set on item")
		}
	})
}

func TestHandleMoveItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		from := createLocation(t, queries, "Garage")
		to := createLocation(t, queries, "Basement")
		item := createItem(t, queries, "Hammer", &from.ID)

		body := strings.NewReader(`{"toLocationId":` + strconv.FormatInt(to.ID, 10) + `}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/items/"+strconv.FormatInt(item.ID, 10)+"/move", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var got models.Item
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.LocationID == nil || *got.LocationID != to.ID {
			t.Errorf("locationId = %v, want %d", got.LocationID, to.ID)
		}
	})
}

func TestHandleGetMoveHistory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		from := createLocation(t, queries, "Garage")
		to := createLocation(t, queries, "Basement")
		item := createItem(t, queries, "Hammer", &from.ID)

		moveBody := strings.NewReader(`{"toLocationId":` + strconv.FormatInt(to.ID, 10) + `}`)
		moveReq := httptest.NewRequest(http.MethodPost, "/v1/items/"+strconv.FormatInt(item.ID, 10)+"/move", moveBody)
		moveReq.Header.Set("Content-Type", "application/json")
		moveRec := httptest.NewRecorder()
		server.ServeHTTP(moveRec, moveReq)
		if moveRec.Code != http.StatusOK {
			t.Fatalf("move status = %d, want %d", moveRec.Code, http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/items/"+strconv.FormatInt(item.ID, 10)+"/history", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var history []models.MoveHistory
		if err := json.NewDecoder(rec.Body).Decode(&history); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(history) != 1 {
			t.Errorf("history len = %d, want %d", len(history), 1)
		}
	})
}

func TestHandleListTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		if _, err := queries.CreateTag("Tools"); err != nil {
			t.Fatalf("create tag: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/v1/tags", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var tags []models.Tag
		if err := json.NewDecoder(rec.Body).Decode(&tags); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(tags) != 1 {
			t.Errorf("tags len = %d, want %d", len(tags), 1)
		}
	})
}

func TestHandleCreateTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, _ := setupTestServer(t)

		body := strings.NewReader(`{"name":"Tools"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/tags", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		if rec.Header().Get("Location") == "" {
			t.Errorf("missing Location header")
		}

		var tag models.Tag
		if err := json.NewDecoder(rec.Body).Decode(&tag); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if tag.Name != "Tools" {
			t.Errorf("name = %q, want %q", tag.Name, "Tools")
		}
	})
}

func TestHandleDeleteTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		tag, err := queries.CreateTag("Tools")
		if err != nil {
			t.Fatalf("create tag: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/v1/tags/"+strconv.FormatInt(tag.ID, 10), nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func TestHandleSearch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, queries := setupTestServer(t)

		createItem(t, queries, "Hammer", nil)
		createItem(t, queries, "Wrench", nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/search?q=hammer", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var result models.SearchResult
		if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(result.Items) == 0 {
			t.Errorf("items len = %d, want > 0", len(result.Items))
		}
	})
}

func TestHandleSearch_MissingQuery(t *testing.T) {
	t.Run("missing_query", func(t *testing.T) {
		server, _ := setupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/v1/search", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleQRCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, _ := setupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/v1/qr/item/1", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("content-type = %q, want %q", ct, "image/png")
		}
	})
}

func TestHandleQRCode_InvalidType(t *testing.T) {
	t.Run("invalid_type", func(t *testing.T) {
		server, _ := setupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/v1/qr/invalid/1", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
