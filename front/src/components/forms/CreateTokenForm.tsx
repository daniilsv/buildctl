import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createToken } from '../../api/tokens'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card'

interface CreateTokenFormProps {
  onSuccess?: () => void
}

export default function CreateTokenForm({ onSuccess }: CreateTokenFormProps) {
  const [name, setName] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [createdToken, setCreatedToken] = useState<string | null>(null)

  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: createToken,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
      setCreatedToken(data.token)
      setName('')
      setExpiresAt('')
      if (onSuccess) {
        setTimeout(onSuccess, 2000)
      }
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const expiresAtUnix = expiresAt ? Math.floor(new Date(expiresAt).getTime() / 1000) : undefined

    mutation.mutate({
      name,
      expires_at: expiresAtUnix,
    })
  }

  if (createdToken) {
    return (
      <Card className="border-green-500">
        <CardHeader>
          <CardTitle className="text-green-600">Token created successfully!</CardTitle>
          <CardDescription>
            <strong>Save this token now - it won't be shown again:</strong>
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="p-4 bg-muted rounded-md font-mono text-sm break-all">
            {createdToken}
          </div>
          <div className="flex gap-2">
            <Button
              onClick={() => {
                navigator.clipboard.writeText(createdToken)
                alert('Token copied to clipboard!')
              }}
            >
              Copy to Clipboard
            </Button>
            <Button variant="outline" onClick={() => setCreatedToken(null)}>
              Create Another
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="token-name">Name *</Label>
        <Input
          id="token-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="expires">Expires At (optional)</Label>
        <Input
          id="expires"
          type="datetime-local"
          value={expiresAt}
          onChange={(e) => setExpiresAt(e.target.value)}
        />
      </div>

      {mutation.isError && (
        <div className="text-sm text-destructive">
          Error: {mutation.error instanceof Error ? mutation.error.message : 'Failed to create token'}
        </div>
      )}

      <div className="flex justify-end">
        <Button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? 'Creating...' : 'Create Token'}
        </Button>
      </div>
    </form>
  )
}
