import type {
  Location,
  Item,
  Tag,
  MoveHistory,
  SearchResult,
  CreateLocationRequest,
  UpdateLocationRequest,
  CreateItemRequest,
  UpdateItemRequest,
  MoveItemRequest,
} from "./types";

const BASE = "/v1";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.message ?? res.statusText);
  }
  return res.json() as Promise<T>;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// --- Locations ---

export async function listLocations(): Promise<Location[]> {
  return request<Location[]>("/locations");
}

export async function getLocation(id: number): Promise<Location> {
  return request<Location>(`/locations/${id}`);
}

export async function createLocation(
  req: CreateLocationRequest,
): Promise<Location> {
  return request<Location>("/locations", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function updateLocation(
  id: number,
  req: UpdateLocationRequest,
): Promise<Location> {
  return request<Location>(`/locations/${id}`, {
    method: "PATCH",
    body: JSON.stringify(req),
  });
}

export async function deleteLocation(id: number): Promise<void> {
  await fetch(`${BASE}/locations/${id}`, { method: "DELETE" });
}

// --- Items ---

export async function listItems(
  locationId?: number,
  tagId?: number,
): Promise<Item[]> {
  const params = new URLSearchParams();
  if (locationId !== undefined) params.set("locationId", String(locationId));
  if (tagId !== undefined) params.set("tagId", String(tagId));
  const qs = params.toString();
  return request<Item[]>(`/items${qs ? `?${qs}` : ""}`);
}

export async function getItem(id: number): Promise<Item> {
  return request<Item>(`/items/${id}`);
}

export async function createItem(req: CreateItemRequest): Promise<Item> {
  return request<Item>("/items", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function updateItem(
  id: number,
  req: UpdateItemRequest,
): Promise<Item> {
  return request<Item>(`/items/${id}`, {
    method: "PATCH",
    body: JSON.stringify(req),
  });
}

export async function deleteItem(id: number): Promise<void> {
  await fetch(`${BASE}/items/${id}`, { method: "DELETE" });
}

export async function uploadPhoto(id: number, file: File): Promise<Item> {
  const form = new FormData();
  form.append("photo", file);
  const res = await fetch(`${BASE}/items/${id}/photo`, {
    method: "POST",
    body: form,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.message ?? res.statusText);
  }
  return res.json() as Promise<Item>;
}

export async function moveItem(
  id: number,
  req: MoveItemRequest,
): Promise<Item> {
  return request<Item>(`/items/${id}/move`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function getMoveHistory(id: number): Promise<MoveHistory[]> {
  return request<MoveHistory[]>(`/items/${id}/history`);
}

// --- Tags ---

export async function listTags(): Promise<Tag[]> {
  return request<Tag[]>("/tags");
}

export async function createTag(name: string): Promise<Tag> {
  return request<Tag>("/tags", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export async function deleteTag(id: number): Promise<void> {
  await fetch(`${BASE}/tags/${id}`, { method: "DELETE" });
}

// --- Search ---

export async function search(query: string): Promise<SearchResult> {
  return request<SearchResult>(`/search?q=${encodeURIComponent(query)}`);
}

// --- QR Code ---

export function qrCodeUrl(type: "item" | "location", id: number): string {
  return `${BASE}/qr/${type}/${id}`;
}
