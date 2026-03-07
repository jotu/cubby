package models

import "time"

type Location struct {
	ID          int64      `json:"id"`
	ParentID    *int64     `json:"parentId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Children    []Location `json:"children,omitempty"`
	ItemCount   int        `json:"itemCount,omitempty"`
}

type CreateLocationRequest struct {
	ParentID    *int64 `json:"parentId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateLocationRequest struct {
	ParentID    *int64  `json:"parentId,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Item struct {
	ID          int64     `json:"id"`
	LocationID  *int64    `json:"locationId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	PhotoPath   string    `json:"photoPath,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Tags        []Tag     `json:"tags,omitempty"`
	Location    *Location `json:"location,omitempty"`
}

type CreateItemRequest struct {
	LocationID  *int64   `json:"locationId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Quantity    int      `json:"quantity"`
	Tags        []string `json:"tags"`
}

type UpdateItemRequest struct {
	LocationID  *int64   `json:"locationId,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type MoveItemRequest struct {
	ToLocationID *int64 `json:"toLocationId"`
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type MoveHistory struct {
	ID             int64     `json:"id"`
	ItemID         int64     `json:"itemId"`
	FromLocationID *int64    `json:"fromLocationId"`
	ToLocationID   *int64    `json:"toLocationId"`
	MovedAt        time.Time `json:"movedAt"`
	FromLocation   *Location `json:"fromLocation,omitempty"`
	ToLocation     *Location `json:"toLocation,omitempty"`
}

type SearchResult struct {
	Items     []Item     `json:"items"`
	Locations []Location `json:"locations"`
}
