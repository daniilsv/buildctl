import { useQuery, useQueryClient } from "@tanstack/react-query";
import { deleteToken, getTokens } from "../api/tokens";
import { deleteSSHKey, getSSHKeys } from "../api/ssh_keys";
import CreateTokenForm from "../components/forms/CreateTokenForm";
import CreateSSHKeyForm from "../components/forms/CreateSSHKeyForm";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";

export default function Settings() {
  const queryClient = useQueryClient();

  const { data: tokens = [] } = useQuery({
    queryKey: ["tokens"],
    queryFn: getTokens,
  });

  const { data: sshKeys = [] } = useQuery({
    queryKey: ["ssh-keys"],
    queryFn: getSSHKeys,
  });

  const handleDelete = async (id: string) => {
    if (confirm("Are you sure you want to delete this token?")) {
      await deleteToken(id);
      queryClient.invalidateQueries({ queryKey: ["tokens"] });
    }
  };

  const handleDeleteSSHKey = async (id: string) => {
    if (confirm("Are you sure you want to delete this SSH key?")) {
      await deleteSSHKey(id);
      queryClient.invalidateQueries({ queryKey: ["ssh-keys"] });
    }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Access Tokens</CardTitle>
          <CardDescription>Manage API tokens for CLI access</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <CreateTokenForm />

          {tokens.length === 0 ? (
            <p className="text-sm text-muted-foreground">No tokens created yet</p>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Last Used</TableHead>
                    <TableHead>Expires</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.map((token) => (
                    <TableRow key={token.id}>
                      <TableCell className="font-medium">{token.name}</TableCell>
                      <TableCell>{new Date(token.created_at).toLocaleString()}</TableCell>
                      <TableCell>
                        {token.last_used_at ? new Date(token.last_used_at).toLocaleString() : 'Never'}
                      </TableCell>
                      <TableCell>
                        {token.expires_at ? new Date(token.expires_at).toLocaleString() : 'Never'}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button variant="destructive" size="sm" onClick={() => handleDelete(token.id)}>
                          Delete
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>SSH Keys</CardTitle>
          <CardDescription>Manage SSH keys for deploy actions on successful builds</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <CreateSSHKeyForm />

          {sshKeys.length === 0 ? (
            <p className="text-sm text-muted-foreground">No SSH keys added yet</p>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Fingerprint</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sshKeys.map((key) => (
                    <TableRow key={key.id}>
                      <TableCell className="font-medium">{key.name}</TableCell>
                      <TableCell className="font-mono text-sm">{key.fingerprint}</TableCell>
                      <TableCell>{new Date(key.created_at).toLocaleString()}</TableCell>
                      <TableCell className="text-right">
                        <Button variant="destructive" size="sm" onClick={() => handleDeleteSSHKey(key.id)}>
                          Delete
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
