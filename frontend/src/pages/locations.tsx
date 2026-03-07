import { useState } from "react";
import { Plus, ChevronRight, Trash2, Pencil, MapPin } from "lucide-react";
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
import {
  useLocations,
  useCreateLocation,
  useUpdateLocation,
  useDeleteLocation,
} from "@/lib/hooks";
import type { Location } from "@/lib/types";

export function LocationsPage() {
  const { data: locations, isLoading } = useLocations();
  const [createOpen, setCreateOpen] = useState(false);

  if (isLoading) {
    return <div className="text-muted-foreground">Loading locations…</div>;
  }

  // Build tree from flat list
  const tree = buildTree(locations ?? []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Locations</h1>
          <p className="text-muted-foreground">Organize your storage spaces</p>
        </div>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add Location
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New Location</DialogTitle>
            </DialogHeader>
            <CreateLocationForm
              locations={locations ?? []}
              onSuccess={() => setCreateOpen(false)}
            />
          </DialogContent>
        </Dialog>
      </div>

      {tree.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <MapPin className="mb-4 h-12 w-12 text-muted-foreground" />
            <h3 className="font-semibold">No locations yet</h3>
            <p className="text-sm text-muted-foreground">
              Create your first location to start organizing
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {tree.map((loc) => (
            <LocationNode
              key={loc.id}
              location={loc}
              allLocations={locations ?? []}
              depth={0}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function LocationNode({
  location,
  allLocations,
  depth,
}: {
  location: Location & { children?: Location[] };
  allLocations: Location[];
  depth: number;
}) {
  const [expanded, setExpanded] = useState(true);
  const [editOpen, setEditOpen] = useState(false);
  const deleteMutation = useDeleteLocation();

  const hasChildren = location.children && location.children.length > 0;

  return (
    <div style={{ marginLeft: depth * 20 }}>
      <Card className="mb-1">
        <CardContent className="flex items-center gap-3 py-3">
          <button
            onClick={() => setExpanded(!expanded)}
            className="flex h-6 w-6 items-center justify-center rounded hover:bg-accent"
          >
            {hasChildren ? (
              <ChevronRight
                className={`h-4 w-4 transition-transform ${expanded ? "rotate-90" : ""}`}
              />
            ) : (
              <div className="h-4 w-4" />
            )}
          </button>

          <MapPin className="h-4 w-4 text-muted-foreground" />

          <div className="flex-1 min-w-0">
            <div className="font-medium">{location.name}</div>
            {location.description && (
              <div className="text-sm text-muted-foreground truncate">
                {location.description}
              </div>
            )}
          </div>

          {location.itemCount !== undefined && location.itemCount > 0 && (
            <Badge variant="secondary">
              {location.itemCount} item{location.itemCount !== 1 ? "s" : ""}
            </Badge>
          )}

          <Dialog open={editOpen} onOpenChange={setEditOpen}>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <Pencil className="h-3.5 w-3.5" />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Edit Location</DialogTitle>
              </DialogHeader>
              <EditLocationForm
                location={location}
                onSuccess={() => setEditOpen(false)}
              />
            </DialogContent>
          </Dialog>

          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-destructive hover:text-destructive"
            onClick={() => {
              if (confirm(`Delete "${location.name}"?`)) {
                deleteMutation.mutate(location.id);
              }
            }}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </CardContent>
      </Card>

      {expanded &&
        hasChildren &&
        location.children!.map((child) => (
          <LocationNode
            key={child.id}
            location={child}
            allLocations={allLocations}
            depth={depth + 1}
          />
        ))}
    </div>
  );
}

function CreateLocationForm({
  locations,
  onSuccess,
}: {
  locations: Location[];
  onSuccess: () => void;
}) {
  const createMutation = useCreateLocation();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [parentId, setParentId] = useState<string>("none");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(
      {
        name,
        description: description || undefined,
        parentId:
          parentId && parentId !== "none" ? Number(parentId) : undefined,
      },
      { onSuccess },
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Garage, Kitchen Pantry"
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Optional description"
          rows={2}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="parent">Parent Location</Label>
        <Select value={parentId} onValueChange={setParentId}>
          <SelectTrigger>
            <SelectValue placeholder="None (top level)" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">None (top level)</SelectItem>
            {locations.map((loc) => (
              <SelectItem key={loc.id} value={String(loc.id)}>
                {loc.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <Button
        type="submit"
        className="w-full"
        disabled={createMutation.isPending}
      >
        {createMutation.isPending ? "Creating…" : "Create Location"}
      </Button>
    </form>
  );
}

function EditLocationForm({
  location,
  onSuccess,
}: {
  location: Location;
  onSuccess: () => void;
}) {
  const updateMutation = useUpdateLocation(location.id);
  const [name, setName] = useState(location.name);
  const [description, setDescription] = useState(location.description ?? "");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(
      {
        name,
        description: description.trim() ? description : undefined,
      },
      { onSuccess },
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="edit-name">Name</Label>
        <Input
          id="edit-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="edit-description">Description</Label>
        <Textarea
          id="edit-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={2}
        />
      </div>
      <Button
        type="submit"
        className="w-full"
        disabled={updateMutation.isPending}
      >
        {updateMutation.isPending ? "Saving…" : "Save Changes"}
      </Button>
    </form>
  );
}

function buildTree(locations: Location[]): Location[] {
  const map = new Map<number, Location & { children: Location[] }>();
  const roots: Location[] = [];

  for (const loc of locations) {
    map.set(loc.id, { ...loc, children: [] });
  }

  for (const loc of locations) {
    const node = map.get(loc.id)!;
    if (loc.parentId && map.has(loc.parentId)) {
      map.get(loc.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  return roots;
}
