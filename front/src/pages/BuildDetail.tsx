import { useQuery } from "@tanstack/react-query";
import { Download } from "lucide-react";
import { useParams } from "react-router-dom";
import { getBuild } from "../api";
import ArtifactsList from "../components/ArtifactsList";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "../components/ui/accordion";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";

export default function BuildDetail() {
  const { id } = useParams<{ id: string }>();
  const { data: build } = useQuery({
    queryKey: ["build", id],
    queryFn: () => getBuild(id!),
    enabled: !!id,
  });

  if (!build)
    return (
      <div className="flex items-center justify-center p-8">Loading...</div>
    );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "success":
        return <Badge variant="success">{status}</Badge>;
      case "failed":
        return <Badge variant="destructive">{status}</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <div className="flex-1">
              {build.project_name && (
                <CardTitle>{build.project_name}</CardTitle>
              )}
              {build.branch_name && (
                <CardDescription className="mt-1">
                  Branch: {build.branch_name}
                </CardDescription>
              )}
              <CardDescription className="mt-2 font-mono">
                Commit: {build.commit_hash.substring(0, 8)}
              </CardDescription>
              {build.commit_message && (
                <CardDescription className="mt-2 break-words max-w-3xl">
                  {build.commit_message}
                </CardDescription>
              )}
              <CardDescription className="mt-2">
                Started: {new Date(build.started_at).toLocaleString()}
                {build.finished_at &&
                  ` • Finished: ${new Date(
                    build.finished_at
                  ).toLocaleString()}`}
              </CardDescription>
            </div>
            {getStatusBadge(build.status)}
          </div>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Build Logs</CardTitle>
        </CardHeader>
        <CardContent>
          {build.logs && build.logs.length > 0 ? (
            <Accordion type="single" className="w-full">
              {build.logs.map((log, index) => (
                <AccordionItem key={log.id} value={`log-${index}`}>
                  <AccordionTrigger value={`log-${index}`}>
                    <div className="flex items-center gap-4 flex-1">
                      <span className="text-sm text-muted-foreground">
                        {new Date(log.created_at).toLocaleTimeString()}
                      </span>
                      {getStatusBadge(log.status)}
                      <span className="text-sm truncate flex-1 text-left">
                        {log.log_message.substring(0, 60)}
                        {log.log_message.length > 60 ? "..." : ""}
                      </span>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent value={`log-${index}`}>
                    <div className="space-y-2 pt-2">
                      <div className="text-sm whitespace-pre-wrap">
                        {log.log_message}
                      </div>
                      {log.artifact_url && (
                        <Button asChild variant="outline" size="sm">
                          <a
                            href={log.artifact_url}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            <Download className="h-4 w-4 mr-2" />
                            Download Artifact
                          </a>
                        </Button>
                      )}
                    </div>
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          ) : (
            <p className="text-sm text-muted-foreground">No logs available</p>
          )}
        </CardContent>
      </Card>

      <ArtifactsList buildId={id!} />
    </div>
  );
}
