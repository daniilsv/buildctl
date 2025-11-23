import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { createBranch } from '../../api/branches'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import TelegramNotificationsList from './TelegramNotificationsList'
import WebhookUrlsList from './WebhookUrlsList'

interface CreateBranchFormProps {
  onSuccess?: () => void
}

export default function CreateBranchForm({ onSuccess }: CreateBranchFormProps) {
  const { id: projectName } = useParams<{ id: string }>()
  const [name, setName] = useState('')
  const [notificationsEnabled, setNotificationsEnabled] = useState(true)
  const [autoDeploy, setAutoDeploy] = useState(false)
  const [telegramNotifications, setTelegramNotifications] = useState<Array<{ chat_id: string; thread_id: string }>>([])
  const [webhookUrls, setWebhookUrls] = useState<string[]>([])

  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (branch: { name: string; settings?: Record<string, any> }) => 
      createBranch(projectName!, branch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches', projectName] })
      setName('')
      setNotificationsEnabled(true)
      setAutoDeploy(false)
      setTelegramNotifications([])
      setWebhookUrls([])
      if (onSuccess) {
        onSuccess()
      }
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const settings: Record<string, any> = {
      notifications_enabled: notificationsEnabled,
      auto_deploy: autoDeploy,
    }

    if (telegramNotifications.length > 0) {
      settings.telegram_notifications = telegramNotifications
        .filter((notif) => notif.chat_id.trim() !== '')
        .map((notif) => {
          const notification: any = {
            chat_id: parseInt(notif.chat_id.trim()),
          }
          if (notif.thread_id.trim() !== '') {
            const threadId = parseInt(notif.thread_id.trim())
            if (!isNaN(threadId)) {
              notification.thread_id = threadId
            }
          }
          return notification
        })
        .filter((notif) => !isNaN(notif.chat_id))
    }

    if (webhookUrls.length > 0) {
      settings.webhook_urls = webhookUrls.filter((url) => url.trim() !== '')
    }

    mutation.mutate({
      name,
      settings,
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

      <TelegramNotificationsList
        value={telegramNotifications}
        onChange={setTelegramNotifications}
      />

      <WebhookUrlsList value={webhookUrls} onChange={setWebhookUrls} />

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
