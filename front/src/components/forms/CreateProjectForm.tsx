import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { createProject } from '../../api'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Select } from '../ui/select'
import TelegramNotificationsList from './TelegramNotificationsList'
import WebhookUrlsList from './WebhookUrlsList'

interface CreateProjectData {
  name: string
  title?: string
  repository_url: string
  repository_type: string
  access_token: string
  settings: Record<string, any>
}

export default function CreateProjectForm() {
  const [name, setName] = useState('')
  const [title, setTitle] = useState('')
  const [repositoryUrl, setRepositoryUrl] = useState('')
  const [repositoryType, setRepositoryType] = useState('gitea')
  const [accessToken, setAccessToken] = useState('')
  const [gitApiUrl, setGitApiUrl] = useState('https://git.int.sktaurus.ru/api/v1')
  const [telegramNotifications, setTelegramNotifications] = useState<Array<{ chat_id: string; thread_id: string }>>([])
  const [webhookUrls, setWebhookUrls] = useState<string[]>([])

  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const mutation = useMutation({
    mutationFn: (project: CreateProjectData) => createProject(project),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      navigate(`/projects/${data.name}`)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const settings: Record<string, any> = {
      git_api_url: gitApiUrl,
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
      title: title || undefined,
      repository_url: repositoryUrl,
      repository_type: repositoryType,
      access_token: accessToken,
      settings,
    } as CreateProjectData)
  }

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
          placeholder="my-project"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="My Project"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="repo-url">Repository URL *</Label>
        <Input
          id="repo-url"
          type="text"
          value={repositoryUrl}
          onChange={(e) => setRepositoryUrl(e.target.value)}
          placeholder="sd/sd-back"
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
        <Label htmlFor="token">Access Token *</Label>
        <Input
          id="token"
          type="password"
          value={accessToken}
          onChange={(e) => setAccessToken(e.target.value)}
          required
        />
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
      />

      <WebhookUrlsList value={webhookUrls} onChange={setWebhookUrls} />

      {mutation.isError && (
        <div className="text-sm text-destructive">
          Error: {mutation.error instanceof Error ? mutation.error.message : 'Failed to create project'}
        </div>
      )}

      <div className="flex justify-end gap-2">
        <Button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? 'Creating...' : 'Create Project'}
        </Button>
      </div>
    </form>
  )
}
