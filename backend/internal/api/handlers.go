package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/joacim/cubby/internal/models"
	"github.com/joacim/cubby/internal/qr"
)

// --- Location Handlers ---

func (s *Server) handleListLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := s.queries.ListLocations()
	if err != nil {
		log.Printf("list locations: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, locations)
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLocationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	loc, err := s.queries.CreateLocation(req)
	if err != nil {
		log.Printf("create location: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.Header().Set("Location", "/v1/locations/"+strconv.FormatInt(loc.ID, 10))
	writeJSON(w, http.StatusCreated, loc)
}

func (s *Server) handleGetLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid location ID")
		return
	}

	loc, err := s.queries.GetLocation(id)
	if err != nil {
		log.Printf("get location: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if loc == nil {
		writeError(w, http.StatusNotFound, "not_found", "Location not found")
		return
	}

	writeJSON(w, http.StatusOK, loc)
}

func (s *Server) handleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid location ID")
		return
	}

	var req models.UpdateLocationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	loc, err := s.queries.UpdateLocation(id, req)
	if err != nil {
		log.Printf("update location: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if loc == nil {
		writeError(w, http.StatusNotFound, "not_found", "Location not found")
		return
	}

	writeJSON(w, http.StatusOK, loc)
}

func (s *Server) handleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid location ID")
		return
	}

	if err := s.queries.DeleteLocation(id); err != nil {
		log.Printf("delete location: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Item Handlers ---

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	var locationID *int64
	var tagID *int64

	if v := r.URL.Query().Get("locationId"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_param", "Invalid locationId")
			return
		}
		locationID = &id
	}
	if v := r.URL.Query().Get("tagId"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_param", "Invalid tagId")
			return
		}
		tagID = &id
	}

	items, err := s.queries.ListItems(locationID, tagID)
	if err != nil {
		log.Printf("list items: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var req models.CreateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	item, err := s.queries.CreateItem(req)
	if err != nil {
		log.Printf("create item: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.Header().Set("Location", "/v1/items/"+strconv.FormatInt(item.ID, 10))
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	item, err := s.queries.GetItem(id)
	if err != nil {
		log.Printf("get item: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "not_found", "Item not found")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	var req models.UpdateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	item, err := s.queries.UpdateItem(id, req)
	if err != nil {
		log.Printf("update item: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "not_found", "Item not found")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	if err := s.queries.DeleteItem(id); err != nil {
		log.Printf("delete item: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	item, err := s.queries.GetItem(id)
	if err != nil {
		log.Printf("get item for photo: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "not_found", "Item not found")
		return
	}

	// Limit upload to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "File too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "Missing photo field")
		return
	}
	defer func() { _ = file.Close() }()

	storedName, err := s.itemService.SavePhoto(id, header.Filename, file)
	if err != nil {
		log.Printf("save photo: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to save photo")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"photoPath": storedName})
}

func (s *Server) handleMoveItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	var req models.MoveItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	item, err := s.itemService.MoveItem(id, req)
	if err != nil {
		log.Printf("move item: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "not_found", "Item not found")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleGetMoveHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid item ID")
		return
	}

	history, err := s.queries.GetMoveHistory(id)
	if err != nil {
		log.Printf("get move history: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, history)
}

// --- Tag Handlers ---

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.queries.ListTags()
	if err != nil {
		log.Printf("list tags: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (s *Server) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	tag, err := s.queries.CreateTag(req.Name)
	if err != nil {
		log.Printf("create tag: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.Header().Set("Location", "/v1/tags/"+strconv.FormatInt(tag.ID, 10))
	writeJSON(w, http.StatusCreated, tag)
}

func (s *Server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid tag ID")
		return
	}

	if err := s.queries.DeleteTag(id); err != nil {
		log.Printf("delete tag: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Search Handler ---

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Query parameter 'q' is required")
		return
	}

	result, err := s.queries.Search(q)
	if err != nil {
		log.Printf("search: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// --- QR Code Handler ---

func (s *Server) handleQRCode(w http.ResponseWriter, r *http.Request) {
	qrType := r.PathValue("type")
	id := r.PathValue("id")

	if qrType != "item" && qrType != "location" {
		writeError(w, http.StatusBadRequest, "invalid_type", "Type must be 'item' or 'location'")
		return
	}

	// Build the URL that the QR code will point to
	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	var path string
	switch qrType {
	case "item":
		path = "/items/" + id
	case "location":
		path = "/locations/" + id
	}

	content := scheme + "://" + host + path

	data, err := qr.Generate(content, 256)
	if err != nil {
		log.Printf("generate qr: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
