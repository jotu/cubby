package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/joacim/cubby/internal/db"
	"github.com/joacim/cubby/internal/models"
)

type ItemService struct {
	queries   *db.Queries
	uploadDir string
}

func NewItemService(q *db.Queries, uploadDir string) *ItemService {
	return &ItemService{queries: q, uploadDir: uploadDir}
}

func (s *ItemService) MoveItem(itemID int64, req models.MoveItemRequest) (*models.Item, error) {
	item, err := s.queries.GetItem(itemID)
	if err != nil {
		return nil, fmt.Errorf("get item for move: %w", err)
	}
	if item == nil {
		return nil, nil
	}

	fromLocationID := item.LocationID

	// Record the move in history
	if err := s.queries.RecordMove(itemID, fromLocationID, req.ToLocationID); err != nil {
		return nil, fmt.Errorf("record move: %w", err)
	}

	// Update the item's location
	updateReq := models.UpdateItemRequest{
		LocationID: req.ToLocationID,
	}
	updated, err := s.queries.UpdateItem(itemID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("update item location: %w", err)
	}

	return updated, nil
}

func (s *ItemService) SavePhoto(itemID int64, filename string, r io.Reader) (string, error) {
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	storedName := fmt.Sprintf("%d_%d%s", itemID, time.Now().UnixNano(), ext)
	fullPath := filepath.Join(s.uploadDir, storedName)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	// Update DB with the relative path
	if err := s.queries.UpdateItemPhoto(itemID, storedName); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("update photo path: %w", err)
	}

	return storedName, nil
}
