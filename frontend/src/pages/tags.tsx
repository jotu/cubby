import { useState } from "react";
import { Plus, Trash2, Tags as TagsIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useTags, useCreateTag, useDeleteTag } from "@/lib/hooks";

export function TagsPage() {
  const { data: tags, isLoading } = useTags();
  const createMutation = useCreateTag();
  const deleteMutation = useDeleteTag();
  const [newTag, setNewTag] = useState("");

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTag.trim()) return;
    createMutation.mutate(newTag.trim(), {
      onSuccess: () => setNewTag(""),
    });
  };

  if (isLoading) {
    return <div className="text-muted-foreground">Loading tags…</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Tags</h1>
        <p className="text-muted-foreground">
          Manage item categories and labels
        </p>
      </div>

      {/* Create form */}
      <form onSubmit={handleCreate} className="flex gap-2">
        <Input
          value={newTag}
          onChange={(e) => setNewTag(e.target.value)}
          placeholder="New tag name…"
          className="max-w-xs"
        />
        <Button
          type="submit"
          disabled={createMutation.isPending || !newTag.trim()}
        >
          <Plus className="mr-2 h-4 w-4" />
          {createMutation.isPending ? "Adding…" : "Add Tag"}
        </Button>
      </form>

      {(tags ?? []).length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <TagsIcon className="mb-4 h-12 w-12 text-muted-foreground" />
            <h3 className="font-semibold">No tags yet</h3>
            <p className="text-sm text-muted-foreground">
              Tags help organize your items into categories
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="flex flex-wrap gap-2">
          {(tags ?? []).map((tag) => (
            <Badge
              key={tag.id}
              variant="secondary"
              className="gap-2 py-1.5 px-3 text-sm"
            >
              {tag.name}
              <button
                onClick={() => {
                  if (confirm(`Delete tag "${tag.name}"?`)) {
                    deleteMutation.mutate(tag.id);
                  }
                }}
                className="rounded-full p-0.5 hover:bg-destructive/20 hover:text-destructive transition-colors"
              >
                <Trash2 className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
