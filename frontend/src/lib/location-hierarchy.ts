import type { Location } from "./types";

export interface LocationOption {
  id: number;
  name: string;
  depth: number;
  path: string;
}

export function buildLocationOptions(locations: Location[]): LocationOption[] {
  const byId = new Map<number, Location>();
  const pathMemo = new Map<number, string>();
  const depthMemo = new Map<number, number>();

  for (const location of locations) {
    byId.set(location.id, location);
  }

  const getPath = (id: number, visited = new Set<number>()): string => {
    const memoized = pathMemo.get(id);
    if (memoized) {
      return memoized;
    }

    const current = byId.get(id);
    if (!current) {
      return "";
    }

    if (visited.has(id)) {
      return current.name;
    }

    visited.add(id);

    if (!current.parentId || !byId.has(current.parentId)) {
      pathMemo.set(id, current.name);
      return current.name;
    }

    const parentPath = getPath(current.parentId, visited);
    const path = `${parentPath} › ${current.name}`;
    pathMemo.set(id, path);
    return path;
  };

  const getDepth = (id: number, visited = new Set<number>()): number => {
    const memoized = depthMemo.get(id);
    if (memoized !== undefined) {
      return memoized;
    }

    const current = byId.get(id);
    if (!current || !current.parentId || !byId.has(current.parentId)) {
      depthMemo.set(id, 0);
      return 0;
    }

    if (visited.has(id)) {
      return 0;
    }

    visited.add(id);
    const depth = getDepth(current.parentId, visited) + 1;
    depthMemo.set(id, depth);
    return depth;
  };

  return locations
    .map((location) => ({
      id: location.id,
      name: location.name,
      depth: getDepth(location.id),
      path: getPath(location.id),
    }))
    .sort((a, b) => a.path.localeCompare(b.path));
}
