// Types matching Go backend models exactly

export interface Location {
  id: number;
  parentId: number | null;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
  children?: Location[];
  itemCount?: number;
}

export interface Item {
  id: number;
  locationId: number | null;
  name: string;
  description: string;
  quantity: number;
  photoPath?: string;
  createdAt: string;
  updatedAt: string;
  tags?: Tag[];
  location?: Location;
}

export interface Tag {
  id: number;
  name: string;
  createdAt: string;
}

export interface MoveHistory {
  id: number;
  itemId: number;
  fromLocationId: number | null;
  toLocationId: number | null;
  movedAt: string;
  fromLocation?: Location;
  toLocation?: Location;
}

export interface SearchResult {
  items: Item[];
  locations: Location[];
}

// Request types

export interface CreateLocationRequest {
  parentId?: number;
  name: string;
  description?: string;
}

export interface UpdateLocationRequest {
  parentId?: number;
  name?: string;
  description?: string;
}

export interface CreateItemRequest {
  locationId?: number;
  name: string;
  description?: string;
  quantity?: number;
  tags?: string[];
}

export interface UpdateItemRequest {
  locationId?: number;
  name?: string;
  description?: string;
  quantity?: number;
  tags?: string[];
}

export interface MoveItemRequest {
  toLocationId: number;
}
