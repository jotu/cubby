import { useMemo, useState } from "react";
import type { ReactNode } from "react";
import { Link } from "react-router";
import {
  ArrowRight,
  FolderSearch,
  MapPin,
  Package,
  Search as SearchIcon,
  Tag,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useItems, useLocations, useSearch, useTags } from "@/lib/hooks";

export function HomePage() {
  const [query, setQuery] = useState("");
  const { data: items, isLoading: itemsLoading } = useItems();
  const { data: locations, isLoading: locationsLoading } = useLocations();
  const { data: tags, isLoading: tagsLoading } = useTags();
  const { data: searchResults, isLoading: searchLoading } = useSearch(query);

  const recentItems = useMemo(
    () =>
      [...(items ?? [])]
        .sort(
          (a, b) =>
            new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
        )
        .slice(0, 5),
    [items],
  );

  const totalItemCount = useMemo(
    () => (items ?? []).reduce((sum, item) => sum + item.quantity, 0),
    [items],
  );

  return (
    <div className="space-y-6">
      <Card className="border-primary/20 bg-gradient-to-br from-primary/10 via-card to-card">
        <CardContent className="space-y-5 p-6">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight">Welcome to Cubby</h1>
            <p className="text-muted-foreground">
              Find things fast, keep storage organized, and jump into your next
              action.
            </p>
          </div>

          <div className="relative">
            <SearchIcon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="h-11 pl-9"
              placeholder="Quick search items and locations…"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>

          <div className="flex flex-wrap gap-2">
            <Button asChild>
              <Link to="/items">
                Open Items
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/locations">Manage Locations</Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/tags">Edit Tags</Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <SummaryCard
          title="Unique Items"
          value={(items ?? []).length}
          loading={itemsLoading}
          icon={<Package className="h-4 w-4" />}
        />
        <SummaryCard
          title="Total Quantity"
          value={totalItemCount}
          loading={itemsLoading}
          icon={<FolderSearch className="h-4 w-4" />}
        />
        <SummaryCard
          title="Locations"
          value={(locations ?? []).length}
          loading={locationsLoading}
          icon={<MapPin className="h-4 w-4" />}
        />
        <SummaryCard
          title="Tags"
          value={(tags ?? []).length}
          loading={tagsLoading}
          icon={<Tag className="h-4 w-4" />}
        />
      </div>

      {query.length >= 2 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Quick Search Results</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {searchLoading && (
              <div className="text-sm text-muted-foreground">Searching…</div>
            )}

            {!searchLoading && searchResults && (
              <div className="space-y-4">
                <div className="space-y-2">
                  <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Items ({searchResults.items.length})
                  </div>
                  {searchResults.items.slice(0, 4).map((item) => (
                    <Link
                      key={item.id}
                      to={`/items/${item.id}`}
                      className="block rounded-lg border p-3 transition-colors hover:bg-accent/50"
                    >
                      <div className="font-medium leading-tight">{item.name}</div>
                      {item.location && (
                        <div className="mt-1 text-xs text-muted-foreground">
                          {item.location.name}
                        </div>
                      )}
                    </Link>
                  ))}
                </div>

                <div className="space-y-2">
                  <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Locations ({searchResults.locations.length})
                  </div>
                  {searchResults.locations.slice(0, 4).map((location) => (
                    <Link
                      key={location.id}
                      to="/locations"
                      className="block rounded-lg border p-3 transition-colors hover:bg-accent/50"
                    >
                      <div className="font-medium leading-tight">{location.name}</div>
                      {location.itemCount !== undefined && location.itemCount > 0 && (
                        <div className="mt-1 text-xs text-muted-foreground">
                          {location.itemCount} item
                          {location.itemCount === 1 ? "" : "s"}
                        </div>
                      )}
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {!searchLoading &&
              query.length >= 2 &&
              searchResults &&
              searchResults.items.length === 0 &&
              searchResults.locations.length === 0 && (
                <div className="text-sm text-muted-foreground">No matches found.</div>
              )}

            <div className="pt-1">
              <Button asChild variant="outline" size="sm">
                <Link to="/search">Open full search</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Recently Updated Items</CardTitle>
        </CardHeader>
        <CardContent>
          {itemsLoading ? (
            <div className="text-sm text-muted-foreground">Loading items…</div>
          ) : recentItems.length === 0 ? (
            <div className="text-sm text-muted-foreground">
              No items yet. Add your first item to get started.
            </div>
          ) : (
            <div className="space-y-2">
              {recentItems.map((item) => (
                <Link
                  key={item.id}
                  to={`/items/${item.id}`}
                  className="flex items-center justify-between rounded-lg border p-3 transition-colors hover:bg-accent/50"
                >
                  <div>
                    <div className="font-medium leading-tight">{item.name}</div>
                    <div className="text-xs text-muted-foreground">
                      Updated {new Date(item.updatedAt).toLocaleDateString()}
                    </div>
                  </div>
                  {item.quantity > 1 && <Badge variant="secondary">×{item.quantity}</Badge>}
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function SummaryCard({
  title,
  value,
  loading,
  icon,
}: {
  title: string;
  value: number;
  loading: boolean;
  icon: ReactNode;
}) {
  return (
    <Card>
      <CardContent className="flex items-center justify-between p-4">
        <div>
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            {title}
          </div>
          <div className="text-2xl font-bold leading-tight">
            {loading ? "…" : value.toLocaleString()}
          </div>
        </div>
        <div className="rounded-full border bg-muted/40 p-2 text-muted-foreground">
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}
