import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getProjects, getBuilds } from "../api";
import CreateProjectForm from "../components/forms/CreateProjectForm";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "../components/ui/dialog";

export default function Dashboard() {
  const [showCreateForm, setShowCreateForm] = useState(false);

  const { data: projects = [] } = useQuery({
    queryKey: ["projects"],
    queryFn: getProjects,
  });

  const { data: builds = [] } = useQuery({
    queryKey: ["builds"],
    queryFn: () => getBuilds(),
  });

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
    <div className="space-y-8">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <Button onClick={() => setShowCreateForm(true)}>
          + Create Project
        </Button>
      </div>

      <Dialog open={showCreateForm} onOpenChange={setShowCreateForm}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create Project</DialogTitle>
          </DialogHeader>
          <div className="mt-4">
            <CreateProjectForm />
          </div>
        </DialogContent>
      </Dialog>

      <div>
        <h2 className="text-2xl font-semibold mb-4">Projects</h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              to={`/projects/${project.name}`}
              className="block"
            >
              <Card className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <CardTitle>{project.title || project.name}</CardTitle>
                  <CardDescription>{project.repository_url}</CardDescription>
                </CardHeader>
              </Card>
            </Link>
          ))}
        </div>
      </div>

      <div>
        <h2 className="text-2xl font-semibold mb-4">Recent Builds</h2>
        <div className="grid gap-4">
          {builds.slice(0, 10).map((build) => (
            <Link key={build.id} to={`/builds/${build.id}`}>
              <Card className="hover:shadow-lg transition-shadow">
                <CardContent className="pt-6">
                  <div className="space-y-2">
                    <div className="flex justify-between items-start gap-4">
                      <div className="flex-1 min-w-0">
                        {build.project_name && (
                          <div className="font-semibold">
                            {build.project_name}
                          </div>
                        )}
                        {build.branch_name && (
                          <div className="text-sm text-muted-foreground">
                            Branch: {build.branch_name}
                          </div>
                        )}
                        {build.commit_message && (
                          <div className="text-sm mt-1 truncate max-w-3xl">
                            {build.commit_message}
                          </div>
                        )}
                      </div>
                      <div className="flex flex-col items-end gap-2 flex-shrink-0">
                        {getStatusBadge(build.status)}
                        <div className="text-xs text-muted-foreground">
                          {new Date(build.started_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">
                      {build.commit_hash.substring(0, 8)}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
