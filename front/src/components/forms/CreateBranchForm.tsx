import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { createBranch } from '../../api/branches'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

interface CreateBranchFormProps {
  onSuccess?: () => void
}

export default function CreateBranchForm({ onSuccess }: CreateBranchFormProps) {
  const { id: projectName } = useParams<{ id: string }>()
  const [name, setName] = useState('')
  const [notificationsEnabled, setNotificationsEnabled] = useState(true)
  const [autoDeploy, setAutoDeploy] = useState(false)

  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (branch: { name: string; settings?: Record<string, any> }) => 
      createBranch(projectName!, branch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectName] })
      setName('')
      setNotificationsEnabled(true)
      setAutoDeploy(false)
      if (onSuccess) {
        onSuccess()
      }
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    mutation.mutate({
      name,
      settings: {
        notifications_enabled: notificationsEnabled,
        auto_deploy: autoDeploy,
      },
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="branch-name">Branch Name *</Label>
        <Input
          id="branch-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="main"
          required
        />
      </div>

      <div className="flex items-center space-x-2">
        <input
          type="checkbox"
          id="notifications"
          checked={notificationsEnabled}
          onChange={(e) => setNotificationsEnabled(e.target.checked)}
          className="h-4 w-4 rounded border-gray-300"
        />
        <Label htmlFor="notifications" className="cursor-pointer">
          Enable Notifications
        </Label>
      </div>

      <div className="flex items-center space-x-2">
        <input
          type="checkbox"
          id="auto-deploy"
          checked={autoDeploy}
          onChange={(e) => setAutoDeploy(e.target.checked)}
          className="h-4 w-4 rounded border-gray-300"
        />
        <Label htmlFor="auto-deploy" className="cursor-pointer">
          Auto Deploy
        </Label>
      </div>

      {mutation.isError && (
        <div className="text-sm text-destructive">
          Error: {mutation.error instanceof Error ? mutation.error.message : 'Failed to create branch'}
        </div>
      )}

      <div className="flex justify-end gap-2">
        <Button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? 'Creating...' : 'Create Branch'}
        </Button>
      </div>
    </form>
  )
}
