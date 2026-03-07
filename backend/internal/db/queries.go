package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/joacim/cubby/internal/models"
)

type Queries struct {
	db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

// --- Locations ---

func (q *Queries) ListLocations() ([]models.Location, error) {
	rows, err := q.db.Query(`
		SELECT l.id, l.parent_id, l.name, l.description, l.created_at, l.updated_at,
			   (SELECT COUNT(*) FROM items WHERE location_id = l.id) as item_count
		FROM locations l
		ORDER BY l.name
	`)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(&loc.ID, &loc.ParentID, &loc.Name, &loc.Description, &loc.CreatedAt, &loc.UpdatedAt, &loc.ItemCount); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

func (q *Queries) GetLocation(id int64) (*models.Location, error) {
	var loc models.Location
	err := q.db.QueryRow(`
		SELECT l.id, l.parent_id, l.name, l.description, l.created_at, l.updated_at,
			   (SELECT COUNT(*) FROM items WHERE location_id = l.id) as item_count
		FROM locations l WHERE l.id = ?
	`, id).Scan(&loc.ID, &loc.ParentID, &loc.Name, &loc.Description, &loc.CreatedAt, &loc.UpdatedAt, &loc.ItemCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get location: %w", err)
	}
	return &loc, nil
}

func (q *Queries) CreateLocation(req models.CreateLocationRequest) (*models.Location, error) {
	result, err := q.db.Exec(`
		INSERT INTO locations (parent_id, name, description) VALUES (?, ?, ?)
	`, req.ParentID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("create location: %w", err)
	}
	id, _ := result.LastInsertId()
	return q.GetLocation(id)
}

func (q *Queries) UpdateLocation(id int64, req models.UpdateLocationRequest) (*models.Location, error) {
	sets := []string{}
	args := []any{}

	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if req.ParentID != nil {
		sets = append(sets, "parent_id = ?")
		args = append(args, *req.ParentID)
	}

	if len(sets) == 0 {
		return q.GetLocation(id)
	}

	sets = append(sets, "updated_at = ?")
	args = append(args, time.Now().UTC())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE locations SET %s WHERE id = ?", strings.Join(sets, ", "))
	if _, err := q.db.Exec(query, args...); err != nil {
		return nil, fmt.Errorf("update location: %w", err)
	}
	return q.GetLocation(id)
}

func (q *Queries) DeleteLocation(id int64) error {
	_, err := q.db.Exec("DELETE FROM locations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete location: %w", err)
	}
	return nil
}

// --- Items ---

func (q *Queries) ListItems(locationID *int64, tagID *int64) ([]models.Item, error) {
	query := `
		SELECT DISTINCT i.id, i.location_id, i.name, i.description, i.quantity, i.photo_path, i.created_at, i.updated_at
		FROM items i
	`
	var args []any
	var wheres []string

	if tagID != nil {
		query += " JOIN item_tags it ON i.id = it.item_id"
		wheres = append(wheres, "it.tag_id = ?")
		args = append(args, *tagID)
	}

	if locationID != nil {
		wheres = append(wheres, "i.location_id = ?")
		args = append(args, *locationID)
	}

	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " ORDER BY i.updated_at DESC"

	rows, err := q.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.LocationID, &item.Name, &item.Description, &item.Quantity, &item.PhotoPath, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load tags for each item
	for i := range items {
		tags, err := q.GetItemTags(items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Tags = tags
	}

	return items, nil
}

func (q *Queries) GetItem(id int64) (*models.Item, error) {
	var item models.Item
	err := q.db.QueryRow(`
		SELECT i.id, i.location_id, i.name, i.description, i.quantity, i.photo_path, i.created_at, i.updated_at
		FROM items i WHERE i.id = ?
	`, id).Scan(&item.ID, &item.LocationID, &item.Name, &item.Description, &item.Quantity, &item.PhotoPath, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}

	tags, err := q.GetItemTags(item.ID)
	if err != nil {
		return nil, err
	}
	item.Tags = tags

	if item.LocationID != nil {
		loc, err := q.GetLocation(*item.LocationID)
		if err != nil {
			return nil, err
		}
		item.Location = loc
	}

	return &item, nil
}

func (q *Queries) CreateItem(req models.CreateItemRequest) (*models.Item, error) {
	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}
	result, err := q.db.Exec(`
		INSERT INTO items (location_id, name, description, quantity) VALUES (?, ?, ?, ?)
	`, req.LocationID, req.Name, req.Description, quantity)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	id, _ := result.LastInsertId()

	// Handle tags
	if len(req.Tags) > 0 {
		if err := q.SetItemTags(id, req.Tags); err != nil {
			return nil, err
		}
	}

	return q.GetItem(id)
}

func (q *Queries) UpdateItem(id int64, req models.UpdateItemRequest) (*models.Item, error) {
	sets := []string{}
	args := []any{}

	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Quantity != nil {
		sets = append(sets, "quantity = ?")
		args = append(args, *req.Quantity)
	}
	if req.LocationID != nil {
		sets = append(sets, "location_id = ?")
		args = append(args, *req.LocationID)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = ?")
		args = append(args, time.Now().UTC())
		args = append(args, id)

		query := fmt.Sprintf("UPDATE items SET %s WHERE id = ?", strings.Join(sets, ", "))
		if _, err := q.db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("update item: %w", err)
		}
	}

	if req.Tags != nil {
		if err := q.SetItemTags(id, req.Tags); err != nil {
			return nil, err
		}
	}

	return q.GetItem(id)
}

func (q *Queries) UpdateItemPhoto(id int64, photoPath string) error {
	_, err := q.db.Exec("UPDATE items SET photo_path = ?, updated_at = ? WHERE id = ?", photoPath, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update item photo: %w", err)
	}
	return nil
}

func (q *Queries) DeleteItem(id int64) error {
	_, err := q.db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

// --- Tags ---

func (q *Queries) ListTags() ([]models.Tag, error) {
	rows, err := q.db.Query("SELECT id, name, created_at FROM tags ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (q *Queries) CreateTag(name string) (*models.Tag, error) {
	result, err := q.db.Exec("INSERT INTO tags (name) VALUES (?) ON CONFLICT (name) DO NOTHING", name)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		// Tag already exists, fetch it
		var tag models.Tag
		err := q.db.QueryRow("SELECT id, name, created_at FROM tags WHERE name = ?", name).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("get existing tag: %w", err)
		}
		return &tag, nil
	}

	var tag models.Tag
	err = q.db.QueryRow("SELECT id, name, created_at FROM tags WHERE id = ?", id).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get created tag: %w", err)
	}
	return &tag, nil
}

func (q *Queries) DeleteTag(id int64) error {
	_, err := q.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}

func (q *Queries) GetItemTags(itemID int64) ([]models.Tag, error) {
	rows, err := q.db.Query(`
		SELECT t.id, t.name, t.created_at
		FROM tags t JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?
		ORDER BY t.name
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("get item tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (q *Queries) SetItemTags(itemID int64, tagNames []string) error {
	tx, err := q.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Remove existing tags
	if _, err := tx.Exec("DELETE FROM item_tags WHERE item_id = ?", itemID); err != nil {
		return fmt.Errorf("clear item tags: %w", err)
	}

	// Add new tags
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Upsert tag
		if _, err := tx.Exec("INSERT INTO tags (name) VALUES (?) ON CONFLICT (name) DO NOTHING", name); err != nil {
			return fmt.Errorf("upsert tag: %w", err)
		}

		// Get tag ID
		var tagID int64
		if err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&tagID); err != nil {
			return fmt.Errorf("get tag id: %w", err)
		}

		// Link
		if _, err := tx.Exec("INSERT INTO item_tags (item_id, tag_id) VALUES (?, ?)", itemID, tagID); err != nil {
			return fmt.Errorf("link item tag: %w", err)
		}
	}

	return tx.Commit()
}

// --- Move History ---

func (q *Queries) RecordMove(itemID int64, fromLocationID, toLocationID *int64) error {
	_, err := q.db.Exec(`
		INSERT INTO move_history (item_id, from_location_id, to_location_id) VALUES (?, ?, ?)
	`, itemID, fromLocationID, toLocationID)
	if err != nil {
		return fmt.Errorf("record move: %w", err)
	}
	return nil
}

func (q *Queries) GetMoveHistory(itemID int64) ([]models.MoveHistory, error) {
	rows, err := q.db.Query(`
		SELECT id, item_id, from_location_id, to_location_id, moved_at
		FROM move_history WHERE item_id = ?
		ORDER BY moved_at DESC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("get move history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []models.MoveHistory
	for rows.Next() {
		var mh models.MoveHistory
		if err := rows.Scan(&mh.ID, &mh.ItemID, &mh.FromLocationID, &mh.ToLocationID, &mh.MovedAt); err != nil {
			return nil, fmt.Errorf("scan move history: %w", err)
		}

		if mh.FromLocationID != nil {
			mh.FromLocation, _ = q.GetLocation(*mh.FromLocationID)
		}
		if mh.ToLocationID != nil {
			mh.ToLocation, _ = q.GetLocation(*mh.ToLocationID)
		}

		history = append(history, mh)
	}
	return history, rows.Err()
}

// --- Search ---

func (q *Queries) Search(query string) (*models.SearchResult, error) {
	result := &models.SearchResult{}

	// Search items via FTS
	rows, err := q.db.Query(`
		SELECT i.id, i.location_id, i.name, i.description, i.quantity, i.photo_path, i.created_at, i.updated_at
		FROM items i
		JOIN items_fts ON items_fts.rowid = i.id
		WHERE items_fts MATCH ?
		ORDER BY rank
	`, query)
	if err != nil {
		return nil, fmt.Errorf("search items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.LocationID, &item.Name, &item.Description, &item.Quantity, &item.PhotoPath, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan search item: %w", err)
		}
		tags, _ := q.GetItemTags(item.ID)
		item.Tags = tags
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Search locations by name
	locRows, err := q.db.Query(`
		SELECT l.id, l.parent_id, l.name, l.description, l.created_at, l.updated_at,
			   (SELECT COUNT(*) FROM items WHERE location_id = l.id) as item_count
		FROM locations l
		WHERE l.name LIKE ? OR l.description LIKE ?
		ORDER BY l.name
	`, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("search locations: %w", err)
	}
	defer func() { _ = locRows.Close() }()

	for locRows.Next() {
		var loc models.Location
		if err := locRows.Scan(&loc.ID, &loc.ParentID, &loc.Name, &loc.Description, &loc.CreatedAt, &loc.UpdatedAt, &loc.ItemCount); err != nil {
			return nil, fmt.Errorf("scan search location: %w", err)
		}
		result.Locations = append(result.Locations, loc)
	}

	return result, locRows.Err()
}
