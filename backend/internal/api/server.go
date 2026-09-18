package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/joacim/cubby/internal/db"
	"github.com/joacim/cubby/internal/service"
)

type Server struct {
	queries     *db.Queries
	itemService *service.ItemService
	uploadDir   string
	mux         *http.ServeMux
}

func NewServer(q *db.Queries, uploadDir string) *Server {
	return NewServerWithFrontendProxy(q, uploadDir, "")
}

func NewServerWithFrontendProxy(q *db.Queries, uploadDir, frontendDevURL string) *Server {
	s := &Server{
		queries:     q,
		itemService: service.NewItemService(q, uploadDir),
		uploadDir:   uploadDir,
		mux:         http.NewServeMux(),
	}
	s.routes(frontendDevURL)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// --- Routes ---

func (s *Server) routes(frontendDevURL string) {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Locations
	s.mux.HandleFunc("GET /v1/locations", s.handleListLocations)
	s.mux.HandleFunc("POST /v1/locations", s.handleCreateLocation)
	s.mux.HandleFunc("GET /v1/locations/{id}", s.handleGetLocation)
	s.mux.HandleFunc("PATCH /v1/locations/{id}", s.handleUpdateLocation)
	s.mux.HandleFunc("DELETE /v1/locations/{id}", s.handleDeleteLocation)

	// Items
	s.mux.HandleFunc("GET /v1/items", s.handleListItems)
	s.mux.HandleFunc("POST /v1/items", s.handleCreateItem)
	s.mux.HandleFunc("GET /v1/items/{id}", s.handleGetItem)
	s.mux.HandleFunc("PATCH /v1/items/{id}", s.handleUpdateItem)
	s.mux.HandleFunc("DELETE /v1/items/{id}", s.handleDeleteItem)
	s.mux.HandleFunc("POST /v1/items/{id}/photo", s.handleUploadPhoto)
	s.mux.HandleFunc("POST /v1/items/{id}/move", s.handleMoveItem)
	s.mux.HandleFunc("GET /v1/items/{id}/history", s.handleGetMoveHistory)

	// Tags
	s.mux.HandleFunc("GET /v1/tags", s.handleListTags)
	s.mux.HandleFunc("POST /v1/tags", s.handleCreateTag)
	s.mux.HandleFunc("DELETE /v1/tags/{id}", s.handleDeleteTag)

	// Search
	s.mux.HandleFunc("GET /v1/search", s.handleSearch)

	// QR code
	s.mux.HandleFunc("GET /v1/qr/{type}/{id}", s.handleQRCode)

	// Uploads
	s.mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(s.uploadDir))))

	// Static frontend (catch-all)
	if frontendDevURL != "" {
		proxyURL, err := url.Parse(frontendDevURL)
		if err != nil {
			log.Printf("invalid frontend dev url %q: %v", frontendDevURL, err)
			s.mux.Handle("GET /", http.FileServer(http.Dir("static")))
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = proxyURL.Host
		}

		s.mux.Handle("GET /", proxy)
		return
	}

	s.mux.Handle("GET /", http.FileServer(http.Dir("static")))
}

// --- JSON Helpers ---

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode json: %v", err)
	}
}

func readJSON(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1 MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Code: code, Message: message})
}

// --- Middleware ---

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// parseID extracts an int64 from a path value.
func parseID(r *http.Request, name string) (int64, error) {
	val := r.PathValue(name)
	var id int64
	_, err := fmt.Sscanf(val, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, val)
	}
	return id, nil
}
