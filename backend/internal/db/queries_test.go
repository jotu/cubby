package db

import (
	"testing"

	"github.com/joacim/cubby/internal/models"
)

func setupTestDB(t *testing.T) *Queries {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	return NewQueries(database)
}

// --- Location Tests ---

func TestCreateLocation(t *testing.T) {
	q := setupTestDB(t)

	tests := []struct {
		name    string
		req     models.CreateLocationRequest
		wantErr bool
	}{
		{
			name: "valid location",
			req:  models.CreateLocationRequest{Name: "Garage"},
		},
		{
			name: "with description",
			req:  models.CreateLocationRequest{Name: "Attic", Description: "Top floor storage"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := q.CreateLocation(tc.req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if loc.Name != tc.req.Name {
				t.Errorf("name = %q, want %q", loc.Name, tc.req.Name)
			}
			if loc.Description != tc.req.Description {
				t.Errorf("description = %q, want %q", loc.Description, tc.req.Description)
			}
			if loc.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if loc.CreatedAt.IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
		})
	}
}

func TestCreateLocation_WithParent(t *testing.T) {
	q := setupTestDB(t)

	parent, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	child, err := q.CreateLocation(models.CreateLocationRequest{
		Name:     "Shelf 1",
		ParentID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	if child.ParentID == nil {
		t.Fatal("expected non-nil ParentID")
	}
	if *child.ParentID != parent.ID {
		t.Errorf("parentID = %d, want %d", *child.ParentID, parent.ID)
	}
}

func TestGetLocation(t *testing.T) {
	q := setupTestDB(t)

	created, err := q.CreateLocation(models.CreateLocationRequest{Name: "Kitchen"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := q.GetLocation(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected location, got nil")
	}
	if got.Name != "Kitchen" {
		t.Errorf("name = %q, want %q", got.Name, "Kitchen")
	}
}

func TestGetLocation_NotFound(t *testing.T) {
	q := setupTestDB(t)

	got, err := q.GetLocation(9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestListLocations(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	_, _ = q.CreateLocation(models.CreateLocationRequest{Name: "Attic"})

	locs, err := q.ListLocations()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(locs) != 2 {
		t.Fatalf("count = %d, want 2", len(locs))
	}
	// Sorted by name
	if locs[0].Name != "Attic" {
		t.Errorf("first = %q, want %q", locs[0].Name, "Attic")
	}
}

func TestListLocations_ItemCount(t *testing.T) {
	q := setupTestDB(t)

	loc, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}

	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Hammer", LocationID: &loc.ID})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Nails", LocationID: &loc.ID})

	locs, err := q.ListLocations()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if locs[0].ItemCount != 2 {
		t.Errorf("itemCount = %d, want 2", locs[0].ItemCount)
	}
}

func TestUpdateLocation(t *testing.T) {
	q := setupTestDB(t)

	loc, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newName := "Workshop"
	updated, err := q.UpdateLocation(loc.ID, models.UpdateLocationRequest{Name: &newName})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Workshop" {
		t.Errorf("name = %q, want %q", updated.Name, "Workshop")
	}
}

func TestUpdateLocation_NoChanges(t *testing.T) {
	q := setupTestDB(t)

	loc, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := q.UpdateLocation(loc.ID, models.UpdateLocationRequest{})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "Garage" {
		t.Errorf("name = %q, want %q", got.Name, "Garage")
	}
}

func TestDeleteLocation(t *testing.T) {
	q := setupTestDB(t)

	loc, err := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := q.DeleteLocation(loc.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, err := q.GetLocation(loc.ID)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

// --- Item Tests ---

func TestCreateItem(t *testing.T) {
	q := setupTestDB(t)

	loc, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})

	tests := []struct {
		name string
		req  models.CreateItemRequest
	}{
		{
			name: "basic item",
			req:  models.CreateItemRequest{Name: "Hammer", LocationID: &loc.ID},
		},
		{
			name: "with quantity",
			req:  models.CreateItemRequest{Name: "Nails", Quantity: 50, LocationID: &loc.ID},
		},
		{
			name: "with tags",
			req:  models.CreateItemRequest{Name: "Wrench", Tags: []string{"tools", "hand tools"}},
		},
		{
			name: "no location",
			req:  models.CreateItemRequest{Name: "Mystery Box"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := q.CreateItem(tc.req)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if item.Name != tc.req.Name {
				t.Errorf("name = %q, want %q", item.Name, tc.req.Name)
			}
			if item.ID == 0 {
				t.Error("expected non-zero ID")
			}
		})
	}
}

func TestCreateItem_DefaultQuantity(t *testing.T) {
	q := setupTestDB(t)

	item, err := q.CreateItem(models.CreateItemRequest{Name: "Laptop"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.Quantity != 1 {
		t.Errorf("quantity = %d, want 1", item.Quantity)
	}
}

func TestGetItem(t *testing.T) {
	q := setupTestDB(t)

	loc, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Office"})
	created, err := q.CreateItem(models.CreateItemRequest{
		Name:       "Laptop",
		LocationID: &loc.ID,
		Tags:       []string{"electronics"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := q.GetItem(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected item, got nil")
	}
	if got.Name != "Laptop" {
		t.Errorf("name = %q, want %q", got.Name, "Laptop")
	}
	if got.Location == nil {
		t.Fatal("expected location to be populated")
	}
	if got.Location.Name != "Office" {
		t.Errorf("location = %q, want %q", got.Location.Name, "Office")
	}
	if len(got.Tags) != 1 || got.Tags[0].Name != "electronics" {
		t.Errorf("tags = %v, want [electronics]", got.Tags)
	}
}

func TestGetItem_NotFound(t *testing.T) {
	q := setupTestDB(t)

	got, err := q.GetItem(9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestListItems(t *testing.T) {
	q := setupTestDB(t)

	loc, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Hammer", LocationID: &loc.ID})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Nails", LocationID: &loc.ID})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Laptop"})

	t.Run("all items", func(t *testing.T) {
		items, err := q.ListItems(nil, nil)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 3 {
			t.Errorf("count = %d, want 3", len(items))
		}
	})

	t.Run("filter by location", func(t *testing.T) {
		items, err := q.ListItems(&loc.ID, nil)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("count = %d, want 2", len(items))
		}
	})
}

func TestListItems_FilterByTag(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Hammer", Tags: []string{"tools"}})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Laptop", Tags: []string{"electronics"}})

	// Get tag ID for "tools"
	tags, _ := q.ListTags()
	var toolsTagID int64
	for _, tag := range tags {
		if tag.Name == "tools" {
			toolsTagID = tag.ID
			break
		}
	}

	items, err := q.ListItems(nil, &toolsTagID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("count = %d, want 1", len(items))
	}
	if items[0].Name != "Hammer" {
		t.Errorf("name = %q, want %q", items[0].Name, "Hammer")
	}
}

func TestUpdateItem(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer", Quantity: 1})

	newName := "Claw Hammer"
	newQty := 3
	updated, err := q.UpdateItem(item.ID, models.UpdateItemRequest{
		Name:     &newName,
		Quantity: &newQty,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Claw Hammer" {
		t.Errorf("name = %q, want %q", updated.Name, "Claw Hammer")
	}
	if updated.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", updated.Quantity)
	}
}

func TestUpdateItem_Tags(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer", Tags: []string{"tools"}})

	newTags := []string{"hand tools", "workshop"}
	updated, err := q.UpdateItem(item.ID, models.UpdateItemRequest{Tags: newTags})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(updated.Tags) != 2 {
		t.Fatalf("tag count = %d, want 2", len(updated.Tags))
	}
	// Tags are sorted by name
	if updated.Tags[0].Name != "hand tools" {
		t.Errorf("tag[0] = %q, want %q", updated.Tags[0].Name, "hand tools")
	}
	if updated.Tags[1].Name != "workshop" {
		t.Errorf("tag[1] = %q, want %q", updated.Tags[1].Name, "workshop")
	}
}

func TestUpdateItemPhoto(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer"})

	if err := q.UpdateItemPhoto(item.ID, "1_12345.jpg"); err != nil {
		t.Fatalf("update photo: %v", err)
	}

	got, _ := q.GetItem(item.ID)
	if got.PhotoPath != "1_12345.jpg" {
		t.Errorf("photoPath = %q, want %q", got.PhotoPath, "1_12345.jpg")
	}
}

func TestDeleteItem(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer"})

	if err := q.DeleteItem(item.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, _ := q.GetItem(item.ID)
	if got != nil {
		t.Error("expected nil after delete")
	}
}

// --- Tag Tests ---

func TestCreateTag(t *testing.T) {
	q := setupTestDB(t)

	tag, err := q.CreateTag("electronics")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if tag.Name != "electronics" {
		t.Errorf("name = %q, want %q", tag.Name, "electronics")
	}
	if tag.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateTag_Duplicate(t *testing.T) {
	q := setupTestDB(t)

	first, _ := q.CreateTag("electronics")
	second, err := q.CreateTag("electronics")
	if err != nil {
		t.Fatalf("create duplicate: %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("duplicate returned different ID: %d vs %d", second.ID, first.ID)
	}
}

func TestListTags(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateTag("electronics")
	_, _ = q.CreateTag("tools")
	_, _ = q.CreateTag("books")

	tags, err := q.ListTags()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("count = %d, want 3", len(tags))
	}
	// Sorted by name
	if tags[0].Name != "books" {
		t.Errorf("first = %q, want %q", tags[0].Name, "books")
	}
}

func TestDeleteTag(t *testing.T) {
	q := setupTestDB(t)

	tag, _ := q.CreateTag("disposable")

	if err := q.DeleteTag(tag.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	tags, _ := q.ListTags()
	if len(tags) != 0 {
		t.Errorf("count = %d, want 0", len(tags))
	}
}

func TestSetItemTags(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer"})

	// Set initial tags
	if err := q.SetItemTags(item.ID, []string{"tools", "workshop"}); err != nil {
		t.Fatalf("set tags: %v", err)
	}

	tags, _ := q.GetItemTags(item.ID)
	if len(tags) != 2 {
		t.Fatalf("count = %d, want 2", len(tags))
	}

	// Replace with new tags
	if err := q.SetItemTags(item.ID, []string{"garage"}); err != nil {
		t.Fatalf("replace tags: %v", err)
	}

	tags, _ = q.GetItemTags(item.ID)
	if len(tags) != 1 {
		t.Fatalf("count = %d, want 1", len(tags))
	}
	if tags[0].Name != "garage" {
		t.Errorf("tag = %q, want %q", tags[0].Name, "garage")
	}
}

func TestSetItemTags_SkipsEmpty(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Hammer"})

	if err := q.SetItemTags(item.ID, []string{"tools", "", "  "}); err != nil {
		t.Fatalf("set tags: %v", err)
	}

	tags, _ := q.GetItemTags(item.ID)
	if len(tags) != 1 {
		t.Errorf("count = %d, want 1 (empty strings skipped)", len(tags))
	}
}

// --- Move History Tests ---

func TestRecordMove(t *testing.T) {
	q := setupTestDB(t)

	locA, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
	locB, _ := q.CreateLocation(models.CreateLocationRequest{Name: "Attic"})
	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Box", LocationID: &locA.ID})

	if err := q.RecordMove(item.ID, &locA.ID, &locB.ID); err != nil {
		t.Fatalf("record move: %v", err)
	}

	history, err := q.GetMoveHistory(item.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("count = %d, want 1", len(history))
	}

	mh := history[0]
	if mh.ItemID != item.ID {
		t.Errorf("itemID = %d, want %d", mh.ItemID, item.ID)
	}
	if mh.FromLocationID == nil || *mh.FromLocationID != locA.ID {
		t.Errorf("fromLocationID = %v, want %d", mh.FromLocationID, locA.ID)
	}
	if mh.ToLocationID == nil || *mh.ToLocationID != locB.ID {
		t.Errorf("toLocationID = %v, want %d", mh.ToLocationID, locB.ID)
	}
	if mh.FromLocation == nil || mh.FromLocation.Name != "Garage" {
		t.Error("expected FromLocation to be populated with Garage")
	}
	if mh.ToLocation == nil || mh.ToLocation.Name != "Attic" {
		t.Error("expected ToLocation to be populated with Attic")
	}
}

func TestGetMoveHistory_Empty(t *testing.T) {
	q := setupTestDB(t)

	item, _ := q.CreateItem(models.CreateItemRequest{Name: "Box"})

	history, err := q.GetMoveHistory(item.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if history != nil {
		t.Errorf("expected nil, got %d entries", len(history))
	}
}

// --- Search Tests ---

func TestSearch_Items(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Blue Hammer", Description: "Heavy duty"})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Red Screwdriver"})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Green Nails"})

	result, err := q.Search("hammer")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Name != "Blue Hammer" {
		t.Errorf("name = %q, want %q", result.Items[0].Name, "Blue Hammer")
	}
}

func TestSearch_ItemsByDescription(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Hammer", Description: "Heavy duty steel"})
	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Feather"})

	result, err := q.Search("steel")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(result.Items))
	}
}

func TestSearch_Locations(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateLocation(models.CreateLocationRequest{Name: "Garage Workshop"})
	_, _ = q.CreateLocation(models.CreateLocationRequest{Name: "Kitchen"})

	result, err := q.Search("garage")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Locations) != 1 {
		t.Fatalf("location count = %d, want 1", len(result.Locations))
	}
	if result.Locations[0].Name != "Garage Workshop" {
		t.Errorf("name = %q, want %q", result.Locations[0].Name, "Garage Workshop")
	}
}

func TestSearch_NoResults(t *testing.T) {
	q := setupTestDB(t)

	_, _ = q.CreateItem(models.CreateItemRequest{Name: "Hammer"})

	result, err := q.Search("nonexistent")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("item count = %d, want 0", len(result.Items))
	}
	if len(result.Locations) != 0 {
		t.Errorf("location count = %d, want 0", len(result.Locations))
	}
}
