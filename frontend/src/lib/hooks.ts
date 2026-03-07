import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as api from "./api";
import type {
  CreateLocationRequest,
  UpdateLocationRequest,
  CreateItemRequest,
  UpdateItemRequest,
  MoveItemRequest,
} from "./types";

// --- Query Keys ---

export const keys = {
  locations: ["locations"] as const,
  location: (id: number) => ["locations", id] as const,
  items: (locationId?: number, tagId?: number) =>
    ["items", { locationId, tagId }] as const,
  item: (id: number) => ["items", id] as const,
  moveHistory: (id: number) => ["items", id, "history"] as const,
  tags: ["tags"] as const,
  search: (q: string) => ["search", q] as const,
};

// --- Locations ---

export function useLocations() {
  return useQuery({
    queryKey: keys.locations,
    queryFn: api.listLocations,
  });
}

export function useLocation(id: number) {
  return useQuery({
    queryKey: keys.location(id),
    queryFn: () => api.getLocation(id),
  });
}

export function useCreateLocation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateLocationRequest) => api.createLocation(req),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.locations }),
  });
}

export function useUpdateLocation(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: UpdateLocationRequest) => api.updateLocation(id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.locations });
      qc.invalidateQueries({ queryKey: keys.location(id) });
    },
  });
}

export function useDeleteLocation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.deleteLocation(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.locations }),
  });
}

// --- Items ---

export function useItems(locationId?: number, tagId?: number) {
  return useQuery({
    queryKey: keys.items(locationId, tagId),
    queryFn: () => api.listItems(locationId, tagId),
  });
}

export function useItem(id: number) {
  return useQuery({
    queryKey: keys.item(id),
    queryFn: () => api.getItem(id),
  });
}

export function useCreateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateItemRequest) => api.createItem(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: keys.locations });
    },
  });
}

export function useUpdateItem(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: UpdateItemRequest) => api.updateItem(id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: keys.item(id) });
      qc.invalidateQueries({ queryKey: keys.locations });
    },
  });
}

export function useDeleteItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.deleteItem(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: keys.locations });
    },
  });
}

export function useUploadPhoto(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => api.uploadPhoto(id, file),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.item(id) });
      qc.invalidateQueries({ queryKey: ["items"] });
    },
  });
}

export function useMoveItem(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: MoveItemRequest) => api.moveItem(id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.item(id) });
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: keys.locations });
      qc.invalidateQueries({ queryKey: keys.moveHistory(id) });
    },
  });
}

export function useMoveHistory(id: number) {
  return useQuery({
    queryKey: keys.moveHistory(id),
    queryFn: () => api.getMoveHistory(id),
  });
}

// --- Tags ---

export function useTags() {
  return useQuery({
    queryKey: keys.tags,
    queryFn: api.listTags,
  });
}

export function useCreateTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => api.createTag(name),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.tags }),
  });
}

export function useDeleteTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => api.deleteTag(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.tags }),
  });
}

// --- Search ---

export function useSearch(query: string) {
  return useQuery({
    queryKey: keys.search(query),
    queryFn: () => api.search(query),
    enabled: query.length >= 2,
  });
}
