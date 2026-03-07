import { useState, useRef } from "react";
import { useParams, useNavigate } from "react-router";
import {
  ArrowLeft,
  Camera,
  MapPin,
  Pencil,
  QrCode,
  Trash2,
  MoveRight,
  Clock,
  Tag,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
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
  useItem,
  useUpdateItem,
  useDeleteItem,
  useUploadPhoto,
  useMoveItem,
  useMoveHistory,
  useLocations,
} from "@/lib/hooks";
import { qrCodeUrl } from "@/lib/api";
import { buildLocationOptions } from "@/lib/location-hierarchy";

export function ItemDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const itemId = Number(id);
  const { data: item, isLoading } = useItem(itemId);
  const deleteMutation = useDeleteItem();

  if (isLoading) {
    return <div className="text-muted-foreground">Loading…</div>;
  }

  if (!item) {
    return <div className="text-muted-foreground">Item not found</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate("/items")}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold tracking-tight">{item.name}</h1>
          {item.location && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <MapPin className="h-3.5 w-3.5" />
              {item.location.name}
            </div>
          )}
        </div>
        <Button
          variant="outline"
          size="icon"
          className="text-destructive hover:text-destructive"
          onClick={() => {
            if (confirm(`Delete "${item.name}"?`)) {
              deleteMutation.mutate(item.id, {
                onSuccess: () => navigate("/items"),
              });
            }
          }}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Left column */}
        <div className="space-y-6">
          {/* Photo */}
          <PhotoSection itemId={itemId} photoPath={item.photoPath} />

          {/* QR Code */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <QrCode className="h-4 w-4" />
                QR Label
              </CardTitle>
            </CardHeader>
            <CardContent className="flex justify-center">
              <img
                src={qrCodeUrl("item", itemId)}
                alt="QR Code"
                className="h-40 w-40"
              />
            </CardContent>
          </Card>
        </div>

        {/* Right column */}
        <div className="space-y-6">
          {/* Details */}
          <EditDetailsSection item={item} />

          {/* Move */}
          <MoveSection itemId={itemId} currentLocationId={item.locationId} />

          {/* History */}
          <HistorySection itemId={itemId} />
        </div>
      </div>
    </div>
  );
}

function PhotoSection({
  itemId,
  photoPath,
}: {
  itemId: number;
  photoPath?: string;
}) {
  const uploadMutation = useUploadPhoto(itemId);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate(file);
    }
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <Camera className="h-4 w-4" />
          Photo
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {photoPath ? (
          <div className="aspect-video overflow-hidden rounded-md bg-muted">
            <img
              src={`/uploads/${photoPath}`}
              alt="Item photo"
              className="h-full w-full object-cover"
            />
          </div>
        ) : (
          <div className="flex aspect-video items-center justify-center rounded-md border-2 border-dashed text-muted-foreground">
            No photo
          </div>
        )}
        <input
          ref={fileRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleFileChange}
        />
        <Button
          variant="outline"
          className="w-full"
          onClick={() => fileRef.current?.click()}
          disabled={uploadMutation.isPending}
        >
          <Camera className="mr-2 h-4 w-4" />
          {uploadMutation.isPending
            ? "Uploading…"
            : photoPath
              ? "Replace Photo"
              : "Upload Photo"}
        </Button>
      </CardContent>
    </Card>
  );
}

function EditDetailsSection({
  item,
}: {
  item: {
    id: number;
    locationId: number | null;
    name: string;
    description: string;
    quantity: number;
    tags?: { id: number; name: string }[];
  };
}) {
  const updateMutation = useUpdateItem(item.id);
  const { data: locations } = useLocations();
  const [editOpen, setEditOpen] = useState(false);
  const [name, setName] = useState(item.name);
  const [description, setDescription] = useState(item.description);
  const [quantity, setQuantity] = useState(String(item.quantity));
  const [locationId, setLocationId] = useState(
    item.locationId ? String(item.locationId) : "none",
  );
  const [tagsInput, setTagsInput] = useState(
    (item.tags ?? []).map((t) => t.name).join(", "),
  );
  const locationOptions = buildLocationOptions(locations ?? []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    updateMutation.mutate(
      {
        name,
        description,
        quantity: Number(quantity) || 1,
        locationId: locationId === "none" ? undefined : Number(locationId),
        tags,
      },
      { onSuccess: () => setEditOpen(false) },
    );
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">Details</CardTitle>
          <Dialog open={editOpen} onOpenChange={setEditOpen}>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <Pencil className="h-3.5 w-3.5" />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Edit Item</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="ed-name">Name</Label>
                  <Input
                    id="ed-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ed-desc">Description</Label>
                  <Textarea
                    id="ed-desc"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={2}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ed-qty">Quantity</Label>
                  <Input
                    id="ed-qty"
                    type="number"
                    min="1"
                    value={quantity}
                    onChange={(e) => setQuantity(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ed-location">Location</Label>
                  <Select value={locationId} onValueChange={setLocationId}>
                    <SelectTrigger id="ed-location">
                      <SelectValue placeholder="None" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">None</SelectItem>
                      {locationOptions.map((location) => (
                        <SelectItem
                          key={location.id}
                          value={String(location.id)}
                        >
                          {location.path}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ed-tags">Tags</Label>
                  <Input
                    id="ed-tags"
                    value={tagsInput}
                    onChange={(e) => setTagsInput(e.target.value)}
                    placeholder="Comma separated"
                  />
                </div>
                <Button
                  type="submit"
                  className="w-full"
                  disabled={updateMutation.isPending}
                >
                  {updateMutation.isPending ? "Saving…" : "Save"}
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {item.description && (
          <p className="text-sm text-muted-foreground">{item.description}</p>
        )}
        <div className="flex items-center gap-4 text-sm">
          <span>
            Quantity: <strong>{item.quantity}</strong>
          </span>
        </div>
        {item.tags && item.tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {item.tags.map((tag) => (
              <Badge key={tag.id} variant="outline" className="gap-1">
                <Tag className="h-3 w-3" />
                {tag.name}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function MoveSection({
  itemId,
  currentLocationId,
}: {
  itemId: number;
  currentLocationId: number | null;
}) {
  const { data: locations } = useLocations();
  const moveMutation = useMoveItem(itemId);
  const [moveOpen, setMoveOpen] = useState(false);
  const [toLocationId, setToLocationId] = useState<string>("");

  const handleMove = (e: React.FormEvent) => {
    e.preventDefault();
    if (!toLocationId) return;
    moveMutation.mutate(
      { toLocationId: Number(toLocationId) },
      { onSuccess: () => setMoveOpen(false) },
    );
  };

  const availableLocations = (locations ?? []).filter(
    (l) => l.id !== currentLocationId,
  );
  const availableLocationOptions = buildLocationOptions(availableLocations);

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <MoveRight className="h-4 w-4" />
          Move Item
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Dialog open={moveOpen} onOpenChange={setMoveOpen}>
          <DialogTrigger asChild>
            <Button variant="outline" className="w-full">
              Move to another location
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Move Item</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleMove} className="space-y-4">
              <div className="space-y-2">
                <Label>Destination</Label>
                <Select value={toLocationId} onValueChange={setToLocationId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select location" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableLocationOptions.map((loc) => (
                      <SelectItem key={loc.id} value={String(loc.id)}>
                        {loc.path}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={moveMutation.isPending || !toLocationId}
              >
                {moveMutation.isPending ? "Moving…" : "Move"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}

function HistorySection({ itemId }: { itemId: number }) {
  const { data: history } = useMoveHistory(itemId);

  if (!history || history.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <Clock className="h-4 w-4" />
          Move History
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {history.map((entry, i) => (
          <div key={entry.id}>
            {i > 0 && <Separator className="my-2" />}
            <div className="flex items-center gap-2 text-sm">
              <span className="text-muted-foreground">
                {entry.fromLocation?.name ?? "Unassigned"}
              </span>
              <MoveRight className="h-3.5 w-3.5 text-muted-foreground" />
              <span>{entry.toLocation?.name ?? "Unassigned"}</span>
            </div>
            <div className="text-xs text-muted-foreground">
              {new Date(entry.movedAt).toLocaleString()}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
