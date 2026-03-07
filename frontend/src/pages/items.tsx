import { useState } from "react";
import { Link } from "react-router";
import { Plus, Package, MapPin, Tag } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useItems, useLocations, useCreateItem } from "@/lib/hooks";
import { buildLocationOptions } from "@/lib/location-hierarchy";
import type { Item, Location } from "@/lib/types";

export function ItemsPage() {
  const [locationFilter, setLocationFilter] = useState<string>("all");
  const locationId =
    locationFilter && locationFilter !== "all"
      ? Number(locationFilter)
      : undefined;
  const { data: items, isLoading } = useItems(locationId);
  const { data: locations } = useLocations();
  const locationOptions = buildLocationOptions(locations ?? []);
  const [createOpen, setCreateOpen] = useState(false);

  if (isLoading) {
    return <div className="text-muted-foreground">Loading items…</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Items</h1>
          <p className="text-muted-foreground">
            {items?.length ?? 0} item{(items?.length ?? 0) !== 1 ? "s" : ""}{" "}
            total
          </p>
        </div>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add Item
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New Item</DialogTitle>
            </DialogHeader>
            <CreateItemForm
              locations={locations ?? []}
              onSuccess={() => setCreateOpen(false)}
            />
          </DialogContent>
        </Dialog>
      </div>

      {/* Filter */}
      <div className="flex gap-2">
        <Select value={locationFilter} onValueChange={setLocationFilter}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="All locations" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All locations</SelectItem>
            {locationOptions.map((loc) => (
              <SelectItem key={loc.id} value={String(loc.id)}>
                {loc.path}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {(items ?? []).length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Package className="mb-4 h-12 w-12 text-muted-foreground" />
            <h3 className="font-semibold">No items found</h3>
            <p className="text-sm text-muted-foreground">
              {locationFilter !== "all"
                ? "Try a different filter"
                : "Add your first item to get started"}
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {(items ?? []).map((item) => (
            <ItemCard key={item.id} item={item} />
          ))}
        </div>
      )}
    </div>
  );
}

function ItemCard({ item }: { item: Item }) {
  return (
    <Link to={`/items/${item.id}`}>
      <Card className="transition-colors hover:bg-accent/50 cursor-pointer">
        <CardContent className="p-4 space-y-2">
          {item.photoPath && (
            <div className="aspect-video rounded-md overflow-hidden bg-muted mb-2">
              <img
                src={`/uploads/${item.photoPath}`}
                alt={item.name}
                className="h-full w-full object-cover"
              />
            </div>
          )}

          <div className="flex items-start justify-between gap-2">
            <h3 className="font-medium leading-tight">{item.name}</h3>
            {item.quantity > 1 && (
              <Badge variant="secondary">×{item.quantity}</Badge>
            )}
          </div>

          {item.description && (
            <p className="text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}

          <div className="flex flex-wrap items-center gap-2 pt-1">
            {item.location && (
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <MapPin className="h-3 w-3" />
                {item.location.name}
              </div>
            )}
            {item.tags &&
              item.tags.map((tag) => (
                <div
                  key={tag.id}
                  className="flex items-center gap-1 text-xs text-muted-foreground"
                >
                  <Tag className="h-3 w-3" />
                  {tag.name}
                </div>
              ))}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

function CreateItemForm({
  locations,
  onSuccess,
}: {
  locations: Location[];
  onSuccess: () => void;
}) {
  const createMutation = useCreateItem();
  const locationOptions = buildLocationOptions(locations);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [locationId, setLocationId] = useState<string>("none");
  const [tagsInput, setTagsInput] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    createMutation.mutate(
      {
        name,
        description: description || undefined,
        quantity: Number(quantity) || 1,
        locationId:
          locationId && locationId !== "none" ? Number(locationId) : undefined,
        tags: tags.length > 0 ? tags : undefined,
      },
      { onSuccess },
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="item-name">Name</Label>
        <Input
          id="item-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Hammer, Extension Cord"
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="item-desc">Description</Label>
        <Textarea
          id="item-desc"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Optional description"
          rows={2}
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="item-qty">Quantity</Label>
          <Input
            id="item-qty"
            type="number"
            min="1"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="item-loc">Location</Label>
          <Select value={locationId} onValueChange={setLocationId}>
            <SelectTrigger>
              <SelectValue placeholder="None" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">None</SelectItem>
              {locationOptions.map((loc) => (
                <SelectItem key={loc.id} value={String(loc.id)}>
                  {loc.path}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="item-tags">Tags</Label>
        <Input
          id="item-tags"
          value={tagsInput}
          onChange={(e) => setTagsInput(e.target.value)}
          placeholder="tools, electronics (comma separated)"
        />
      </div>
      <Button
        type="submit"
        className="w-full"
        disabled={createMutation.isPending}
      >
        {createMutation.isPending ? "Creating…" : "Create Item"}
      </Button>
    </form>
  );
}
