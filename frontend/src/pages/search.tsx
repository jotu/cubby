import { useState } from "react";
import { Link } from "react-router";
import { Search as SearchIcon, MapPin, Package, Tag } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useSearch } from "@/lib/hooks";

export function SearchPage() {
  const [query, setQuery] = useState("");
  const { data: results, isLoading } = useSearch(query);

  const hasResults =
    results && (results.items.length > 0 || results.locations.length > 0);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Search</h1>
        <p className="text-muted-foreground">Find items and locations</p>
      </div>

      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          className="pl-9"
          placeholder="Search items and locations…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          autoFocus
        />
      </div>

      {isLoading && query.length >= 2 && (
        <div className="text-muted-foreground">Searching…</div>
      )}

      {query.length >= 2 && !isLoading && !hasResults && (
        <div className="text-center py-12 text-muted-foreground">
          <SearchIcon className="mx-auto mb-4 h-12 w-12" />
          <p>No results for &ldquo;{query}&rdquo;</p>
        </div>
      )}

      {query.length > 0 && query.length < 2 && (
        <div className="text-sm text-muted-foreground">
          Type at least 2 characters to search
        </div>
      )}

      {hasResults && (
        <div className="space-y-6">
          {/* Items */}
          {results.items.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
                Items ({results.items.length})
              </h2>
              <div className="space-y-2">
                {results.items.map((item) => (
                  <Link key={item.id} to={`/items/${item.id}`}>
                    <Card className="transition-colors hover:bg-accent/50 cursor-pointer">
                      <CardContent className="flex items-center gap-3 py-3">
                        <Package className="h-4 w-4 text-muted-foreground shrink-0" />
                        <div className="flex-1 min-w-0">
                          <div className="font-medium">{item.name}</div>
                          {item.description && (
                            <div className="text-sm text-muted-foreground truncate">
                              {item.description}
                            </div>
                          )}
                        </div>
                        {item.location && (
                          <div className="flex items-center gap-1 text-xs text-muted-foreground">
                            <MapPin className="h-3 w-3" />
                            {item.location.name}
                          </div>
                        )}
                        {item.quantity > 1 && (
                          <Badge variant="secondary">×{item.quantity}</Badge>
                        )}
                        {item.tags &&
                          item.tags.map((tag) => (
                            <Badge
                              key={tag.id}
                              variant="outline"
                              className="gap-1"
                            >
                              <Tag className="h-3 w-3" />
                              {tag.name}
                            </Badge>
                          ))}
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {/* Locations */}
          {results.locations.length > 0 && (
            <div className="space-y-2">
              <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
                Locations ({results.locations.length})
              </h2>
              <div className="space-y-2">
                {results.locations.map((loc) => (
                  <Link key={loc.id} to="/locations">
                    <Card className="transition-colors hover:bg-accent/50 cursor-pointer">
                      <CardContent className="flex items-center gap-3 py-3">
                        <MapPin className="h-4 w-4 text-muted-foreground shrink-0" />
                        <div className="flex-1 min-w-0">
                          <div className="font-medium">{loc.name}</div>
                          {loc.description && (
                            <div className="text-sm text-muted-foreground truncate">
                              {loc.description}
                            </div>
                          )}
                        </div>
                        {loc.itemCount !== undefined && loc.itemCount > 0 && (
                          <Badge variant="secondary">
                            {loc.itemCount} item{loc.itemCount !== 1 ? "s" : ""}
                          </Badge>
                        )}
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
