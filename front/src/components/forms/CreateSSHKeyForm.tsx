import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createSSHKey } from '../../api/ssh_keys'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

interface CreateSSHKeyFormProps {
  onSuccess?: () => void
}

export default function CreateSSHKeyForm({ onSuccess }: CreateSSHKeyFormProps) {
  const [name, setName] = useState('')
  const [privateKey, setPrivateKey] = useState('')

  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: createSSHKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ssh-keys'] })
      setName('')
      setPrivateKey('')
      if (onSuccess) {
        setTimeout(onSuccess, 1000)
      }
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate({ name, private_key: privateKey })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="ssh-key-name">Name *</Label>
        <Input
          id="ssh-key-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="deploy-key"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="ssh-private-key">Private Key *</Label>
        <textarea
          id="ssh-private-key"
          className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
          value={privateKey}
          onChange={(e) => setPrivateKey(e.target.value)}
          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
          required
        />
      </div>

      {mutation.isError && (
        <div className="text-sm text-destructive">
          Error: {mutation.error instanceof Error ? mutation.error.message : 'Failed to create SSH key'}
        </div>
      )}

      <div className="flex justify-end">
        <Button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? 'Adding...' : 'Add SSH Key'}
        </Button>
      </div>
    </form>
  )
}
