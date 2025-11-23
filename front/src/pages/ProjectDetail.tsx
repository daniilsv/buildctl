import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { deleteProject, getBuilds, getProjectByName, updateProject } from "../api";
import { getBranches, updateBranch } from "../api/branches";
import CreateBranchForm from "../components/forms/CreateBranchForm";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "../components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Select } from "../components/ui/select";
import TelegramNotificationsList from "../components/forms/TelegramNotificationsList";
import WebhookUrlsList from "../components/forms/WebhookUrlsList";

export default function ProjectDetail() {
  const { id: projectName } = useParams<{ id: string }>();
  const [showCreateBranch, setShowCreateBranch] = useState(false);
  const [showEditForm, setShowEditForm] = useState(false);
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { data: project } = useQuery({
    queryKey: ["project", projectName],
    queryFn: () => getProjectByName(projectName!),
    enabled: !!projectName,
  });

  const { data: builds = [] } = useQuery({
    queryKey: ["builds", projectName],
    queryFn: () => getBuilds(projectName),
    enabled: !!projectName,
  });

  const { data: branches = [] } = useQuery({
    queryKey: ["branches", projectName],
    queryFn: () => getBranches(projectName!),
    enabled: !!projectName,
  });

  const handleDelete = async () => {
    if (
      confirm(
        "Are you sure you want to delete this project? This will delete all associated builds and branches."
      )
    ) {
      await deleteProject(project!.name);
      queryClient.invalidateQueries({ queryKey: ["projects"] });
      navigate("/");
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return <Badge variant="success">{status}</Badge>
      case 'failed':
        return <Badge variant="destructive">{status}</Badge>
      default:
        return <Badge variant="secondary">{status}</Badge>
    }
  }

  if (!project) return <div className="flex items-center justify-center p-8">Loading...</div>;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold">{project.title || project.name}</h1>
          <p className="text-muted-foreground mt-1">Repository: {project.repository_url}</p>
          <p className="text-sm text-muted-foreground">Type: {project.repository_type}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setShowEditForm(true)}>
            Edit
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            Delete
          </Button>
        </div>
      </div>

      <Dialog open={showEditForm} onOpenChange={setShowEditForm}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Project</DialogTitle>
          </DialogHeader>
          <EditProjectForm project={project} onSuccess={() => setShowEditForm(false)} />
        </DialogContent>
      </Dialog>

      <Tabs defaultValue="branches" className="w-full">
        <TabsList>
          <TabsTrigger value="branches">Branches</TabsTrigger>
          <TabsTrigger value="builds">Builds</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="branches" className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-semibold">Branches</h2>
            <Button onClick={() => setShowCreateBranch(true)}>
              + Create Branch
            </Button>
          </div>

          <Dialog open={showCreateBranch} onOpenChange={setShowCreateBranch}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Branch</DialogTitle>
              </DialogHeader>
              <CreateBranchForm onSuccess={() => setShowCreateBranch(false)} />
            </DialogContent>
          </Dialog>

          <div className="grid gap-4 md:grid-cols-2">
            {branches.map((branch) => (
              <Card key={branch.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle>{branch.name}</CardTitle>
                      {branch.last_successful_commit && (
                        <CardDescription className="mt-1">
                          Last successful: {branch.last_successful_commit.substring(0, 8)}
                          {branch.last_successful_at &&
                            ` at ${new Date(branch.last_successful_at).toLocaleString()}`}
                        </CardDescription>
                      )}
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        asChild
                      >
                        <Link to={`/projects/${projectName}/branches/${branch.name}`}>
                          View Builds
                        </Link>
                      </Button>
                      <EditBranchButton projectName={projectName!} branchName={branch.name} branch={branch} />
                    </div>
                  </div>
                </CardHeader>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="builds" className="space-y-4">
          <h2 className="text-2xl font-semibold">Builds</h2>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Branch</TableHead>
                  <TableHead>Commit</TableHead>
                  <TableHead>Message</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Started</TableHead>
                  <TableHead>Finished</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {builds.map((build) => (
                  <TableRow key={build.id}>
                    <TableCell>
                      {build.branch_name || '-'}
                    </TableCell>
                    <TableCell>
                      <Link to={`/builds/${build.id}`} className="font-mono text-sm hover:underline">
                        {build.commit_hash.substring(0, 8)}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-xs truncate">
                      {build.commit_message || '-'}
                    </TableCell>
                    <TableCell>{getStatusBadge(build.status)}</TableCell>
                    <TableCell>{new Date(build.started_at).toLocaleString()}</TableCell>
                    <TableCell>
                      {build.finished_at ? new Date(build.finished_at).toLocaleString() : '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </TabsContent>

        <TabsContent value="settings" className="space-y-4">
          <h2 className="text-2xl font-semibold">Project Settings</h2>
          <Card>
            <CardHeader>
              <CardTitle>Repository Settings</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div>
                <Label>Repository URL</Label>
                <p className="text-sm text-muted-foreground">{project.repository_url}</p>
              </div>
              <div>
                <Label>Repository Type</Label>
                <p className="text-sm text-muted-foreground">{project.repository_type}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function EditProjectForm({
  project,
  onSuccess,
}: {
  project: any;
  onSuccess: () => void;
}) {
  const [name, setName] = useState(project.name);
  const [title, setTitle] = useState(project.title || project.name);
  const [repositoryUrl, setRepositoryUrl] = useState(project.repository_url);
  const [repositoryType, setRepositoryType] = useState(project.repository_type);
  const [gitApiUrl, setGitApiUrl] = useState(
    project.settings?.git_api_url || "https://git.int.sktaurus.ru/api/v1"
  );
  const [telegramNotifications, setTelegramNotifications] = useState<
    Array<{ chat_id: string; thread_id: string }>
  >(() => {
    if (project.settings?.telegram_notifications) {
      return project.settings.telegram_notifications.map((notif: any) => ({
        chat_id: String(notif.chat_id || ""),
        thread_id: notif.thread_id ? String(notif.thread_id) : "",
      }));
    }
    return [];
  });
  const [webhookUrls, setWebhookUrls] = useState<string[]>(
    project.settings?.webhook_urls || []
  );
  const queryClient = useQueryClient();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const settings: Record<string, any> = {
      git_api_url: gitApiUrl,
    };

    if (telegramNotifications.length > 0) {
      settings.telegram_notifications = telegramNotifications
        .filter((notif) => notif.chat_id.trim() !== "")
        .map((notif) => {
          const notification: any = {
            chat_id: notif.chat_id.trim(),
          };
          if (notif.thread_id.trim() !== "") {
            notification.thread_id = notif.thread_id.trim();
          }
          return notification;
        });
    }

    if (webhookUrls.length > 0) {
      settings.webhook_urls = webhookUrls.filter((url) => url.trim() !== "");
    }

    await updateProject(project.name, {
      name,
      title,
      repository_url: repositoryUrl,
      repository_type: repositoryType,
      settings,
    });

    queryClient.invalidateQueries({ queryKey: ["project", project.name] });
    onSuccess();
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Name (slug) *</Label>
        <Input
          id="name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="title">Title *</Label>
        <Input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="repo-url">Repository URL *</Label>
        <Input
          id="repo-url"
          type="text"
          value={repositoryUrl}
          onChange={(e) => setRepositoryUrl(e.target.value)}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="repo-type">Repository Type *</Label>
        <Select
          id="repo-type"
          value={repositoryType}
          onChange={(e) => setRepositoryType(e.target.value)}
        >
          <option value="gitea">Gitea</option>
          <option value="git_cli">Git CLI</option>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="git-api">Git API URL</Label>
        <Input
          id="git-api"
          type="text"
          value={gitApiUrl}
          onChange={(e) => setGitApiUrl(e.target.value)}
        />
      </div>

      <TelegramNotificationsList
        value={telegramNotifications}
        onChange={setTelegramNotifications}
        projectName={project.name}
      />

      <WebhookUrlsList 
        value={webhookUrls} 
        onChange={setWebhookUrls}
        projectName={project.name}
      />

      <DialogFooter>
        <Button type="button" variant="outline" onClick={onSuccess}>
          Cancel
        </Button>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
}

function EditBranchButton({ projectName, branchName, branch }: { projectName: string; branchName: string; branch: any }) {
  const [open, setOpen] = useState(false);
  const [telegramNotifications, setTelegramNotifications] = useState<
    Array<{ chat_id: string; thread_id: string }>
  >(() => {
    if (branch.settings?.telegram_notifications) {
      return branch.settings.telegram_notifications.map((notif: any) => ({
        chat_id: String(notif.chat_id || ""),
        thread_id: notif.thread_id ? String(notif.thread_id) : "",
      }));
    }
    return [];
  });
  const [webhookUrls, setWebhookUrls] = useState<string[]>(
    branch.settings?.webhook_urls || []
  );
  const queryClient = useQueryClient();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const settings: Record<string, any> = {};

    if (telegramNotifications.length > 0) {
      settings.telegram_notifications = telegramNotifications
        .filter((notif) => notif.chat_id.trim() !== "")
        .map((notif) => {
          const notification: any = {
            chat_id: notif.chat_id.trim(),
          };
          if (notif.thread_id.trim() !== "") {
            notification.thread_id = notif.thread_id.trim();
          }
          return notification;
        });
    }

    if (webhookUrls.length > 0) {
      settings.webhook_urls = webhookUrls.filter((url) => url.trim() !== "");
    }

    await updateBranch(projectName, branchName, { settings });
    queryClient.invalidateQueries({ queryKey: ["branches", projectName] });
    setOpen(false);
  };

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        Edit
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Branch Settings</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <TelegramNotificationsList
              value={telegramNotifications}
              onChange={setTelegramNotifications}
              projectName={projectName}
              branchName={branchName}
            />

            <WebhookUrlsList 
              value={webhookUrls} 
              onChange={setWebhookUrls}
              projectName={projectName}
              branchName={branchName}
            />

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit">Save</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
