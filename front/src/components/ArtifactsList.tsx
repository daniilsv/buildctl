import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FileText, Package, Trash2 } from "lucide-react";
import { useState } from "react";
import {
  deleteArtifact,
  deleteBuildArtifacts,
  getBuildArtifacts,
  type Artifact,
} from "../api";
import { Button } from "./ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./ui/table";

interface ArtifactsListProps {
  buildId: string;
}

export default function ArtifactsList({ buildId }: ArtifactsListProps) {
  const queryClient = useQueryClient();
  const [deleteAllOpen, setDeleteAllOpen] = useState(false);
  const [deleteArtifactId, setDeleteArtifactId] = useState<string | null>(null);

  const { data: artifacts, isLoading } = useQuery({
    queryKey: ["artifacts", buildId],
    queryFn: () => getBuildArtifacts(buildId),
    enabled: !!buildId,
  });

  const deleteArtifactMutation = useMutation({
    mutationFn: deleteArtifact,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["artifacts", buildId] });
      setDeleteArtifactId(null);
    },
  });

  const deleteAllMutation = useMutation({
    mutationFn: () => deleteBuildArtifacts(buildId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["artifacts", buildId] });
      setDeleteAllOpen(false);
    },
  });

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  };

  const getArtifactIcon = (artifact: Artifact) => {
    if (artifact.artifact_type === "container_image") {
      return <Package className="h-4 w-4" />;
    }
    return <FileText className="h-4 w-4" />;
  };

  const getArtifactDisplayName = (artifact: Artifact): string => {
    if (artifact.artifact_type === "container_image") {
      return artifact.image_name || artifact.filename;
    }
    return artifact.filename;
  };

  const getArtifactDetails = (artifact: Artifact): string => {
    if (artifact.artifact_type === "container_image") {
      const parts = [];
      if (artifact.image_tag) parts.push(artifact.image_tag);
      if (artifact.image_digest)
        parts.push(artifact.image_digest.substring(0, 12));
      return parts.join(" • ");
    }
    return formatBytes(artifact.size_bytes);
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Build Artifacts</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Loading...</p>
        </CardContent>
      </Card>
    );
  }

  if (!artifacts || artifacts.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Build Artifacts</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No artifacts available
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex justify-between items-center">
          <div>
            <CardTitle>Build Artifacts</CardTitle>
            <CardDescription className="mt-1">
              {artifacts.length} artifact{artifacts.length !== 1 ? "s" : ""}
            </CardDescription>
          </div>
          {artifacts.length > 0 && (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setDeleteAllOpen(true)}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete All
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12"></TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Details</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {artifacts.map((artifact) => (
              <TableRow key={artifact.id}>
                <TableCell>{getArtifactIcon(artifact)}</TableCell>
                <TableCell className="font-medium">
                  {getArtifactDisplayName(artifact)}
                </TableCell>
                <TableCell className="text-muted-foreground text-sm">
                  {getArtifactDetails(artifact)}
                </TableCell>
                <TableCell className="text-muted-foreground text-sm">
                  {new Date(artifact.created_at).toLocaleString()}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    {artifact.public_url && (
                      <Button asChild variant="outline" size="sm">
                        <a
                          href={artifact.public_url}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          Download
                        </a>
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setDeleteArtifactId(artifact.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>

      <Dialog open={deleteAllOpen} onOpenChange={setDeleteAllOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete All Artifacts</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete all {artifacts.length} artifact
              {artifacts.length !== 1 ? "s" : ""}? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="destructive"
              onClick={() => deleteAllMutation.mutate()}
              disabled={deleteAllMutation.isPending}
            >
              {deleteAllMutation.isPending ? "Deleting..." : "Delete All"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {deleteArtifactId && (
        <Dialog
          open={!!deleteArtifactId}
          onOpenChange={() => setDeleteArtifactId(null)}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Artifact</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete "
                {getArtifactDisplayName(
                  artifacts.find((a) => a.id === deleteArtifactId)!
                )}
                "? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="destructive"
                onClick={() => deleteArtifactMutation.mutate(deleteArtifactId)}
                disabled={deleteArtifactMutation.isPending}
              >
                {deleteArtifactMutation.isPending ? "Deleting..." : "Delete"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </Card>
  );
}
