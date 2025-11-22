import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { getBuildsByBranch } from '../api/branches'
import { Card, CardContent } from '../components/ui/card'
import { Badge } from '../components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table'

export default function BranchDetail() {
  const { id: projectName, branchName } = useParams<{ id: string; branchName: string }>()

  const { data: builds = [] } = useQuery({
    queryKey: ['builds', projectName, branchName],
    queryFn: () => getBuildsByBranch(projectName!, branchName!),
    enabled: !!projectName && !!branchName,
  })

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

  return (
    <div className="space-y-6">
      <div>
        <Link to={`/projects/${projectName}`} className="text-sm text-muted-foreground hover:underline">
          ← Back to Project
        </Link>
        <h1 className="text-3xl font-bold mt-2">Branch: {branchName}</h1>
      </div>

      <Card>
        <CardContent className="pt-6">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Commit</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Started</TableHead>
                <TableHead>Finished</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {builds.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    No builds found
                  </TableCell>
                </TableRow>
              ) : (
                builds.map((build: any) => (
                  <TableRow key={build.id}>
                    <TableCell>
                      <Link to={`/builds/${build.id}`} className="font-mono text-sm hover:underline">
                        {build.commit_hash.substring(0, 8)}
                      </Link>
                    </TableCell>
                    <TableCell>{getStatusBadge(build.status)}</TableCell>
                    <TableCell>{new Date(build.started_at).toLocaleString()}</TableCell>
                    <TableCell>
                      {build.finished_at ? new Date(build.finished_at).toLocaleString() : '-'}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
