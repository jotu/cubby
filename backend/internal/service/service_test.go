package service

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joacim/cubby/internal/db"
	"github.com/joacim/cubby/internal/models"
)

func setupTestService(t *testing.T) (*ItemService, *db.Queries) {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	uploadDir := t.TempDir()
	q := db.NewQueries(database)
	return NewItemService(q, uploadDir), q
}

func TestMoveItem(t *testing.T) {
	svc, q := setupTestService(t)

	locA, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create location A: %v", err)
	}
	locB, err := q.CreateLocation(models.CreateLocationRequest{Name: "Attic"})
	if err != nil {
		t.Fatalf("create location B: %v", err)
	}

	item, err := q.CreateItem(models.CreateItemRequest{Name: "Box", LocationID: &locA.ID})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	moved, err := svc.MoveItem(item.ID, models.MoveItemRequest{ToLocationID: &locB.ID})
	if err != nil {
		t.Fatalf("move item: %v", err)
	}
	if moved == nil {
		t.Fatal("expected item, got nil")
	}
	if moved.LocationID == nil || *moved.LocationID != locB.ID {
		t.Errorf("locationID = %v, want %d", moved.LocationID, locB.ID)
	}

	// Verify move history was recorded
	history, err := q.GetMoveHistory(item.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history count = %d, want 1", len(history))
	}
	if history[0].FromLocationID == nil || *history[0].FromLocationID != locA.ID {
		t.Errorf("fromLocationID = %v, want %d", history[0].FromLocationID, locA.ID)
	}
	if history[0].ToLocationID == nil || *history[0].ToLocationID != locB.ID {
		t.Errorf("toLocationID = %v, want %d", history[0].ToLocationID, locB.ID)
	}
}

func TestMoveItem_NotFound(t *testing.T) {
	svc, _ := setupTestService(t)

	locID := int64(1)
	moved, err := svc.MoveItem(9999, models.MoveItemRequest{ToLocationID: &locID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if moved != nil {
		t.Errorf("expected nil for non-existent item, got %+v", moved)
	}
}

func TestMoveItem_MultipleMoves(t *testing.T) {
	svc, q := setupTestService(t)

	locA, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	locB, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Attic"})
	locC, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Kitchen"})
	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Box", LocationID: &locA.ID})

	if _, err := svc.MoveItem(item.ID, models.MoveItemRequest{ToLocationID: &locB.ID}); err != nil {
		t.Fatalf("first move: %v", err)
	}
	moved, err := svc.MoveItem(item.ID, models.MoveItemRequest{ToLocationID: &locC.ID})
	if err != nil {
		t.Fatalf("second move: %v", err)
	}
	if moved.LocationID == nil || *moved.LocationID != locC.ID {
		t.Errorf("locationID = %v, want %d", moved.LocationID, locC.ID)
	}

	history, err := q.GetMoveHistory(item.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("history count = %d, want 2", len(history))
	}
}

func TestSavePhoto(t *testing.T) {
	svc, q := setupTestService(t)

	item, err := q.CreateItem(models.CreateItemRequest{Name: "Lamp"})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	content := []byte("fake image data")
	storedName, err := svc.SavePhoto(item.ID, "photo.jpg", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("save photo: %v", err)
	}
	if storedName == "" {
		t.Fatal("expected non-empty stored name")
	}
	if !strings.HasSuffix(storedName, ".jpg") {
		t.Errorf("storedName = %q, want suffix .jpg", storedName)
	}

	// Verify file exists on disk
	fullPath := filepath.Join(svc.uploadDir, storedName)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("file content = %q, want %q", data, content)
	}
}

func TestSavePhoto_DefaultExtension(t *testing.T) {
	svc, q := setupTestService(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Lamp"})

	storedName, err := svc.SavePhoto(item.ID, "noext", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("save photo: %v", err)
	}
	if !strings.HasSuffix(storedName, ".jpg") {
		t.Errorf("storedName = %q, want default .jpg extension", storedName)
	}
}

func TestSavePhoto_PreservesExtension(t *testing.T) {
	svc, q := setupTestService(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Lamp"})

	storedName, err := svc.SavePhoto(item.ID, "image.png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("save photo: %v", err)
	}
	if !strings.HasSuffix(storedName, ".png") {
		t.Errorf("storedName = %q, want suffix .png", storedName)
	}
}

func TestSavePhoto_UpdatesDB(t *testing.T) {
	svc, q := setupTestService(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Lamp"})

	storedName, err := svc.SavePhoto(item.ID, "photo.jpg", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("save photo: %v", err)
	}

	got, err := q.GetItem(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if got.PhotoPath != storedName {
		t.Errorf("photoPath = %q, want %q", got.PhotoPath, storedName)
	}
}
