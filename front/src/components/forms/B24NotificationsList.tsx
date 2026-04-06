import { useState, useEffect } from 'react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { testProjectB24Notification, testBranchB24Notification } from '../../api'

interface B24Notification {
  type_key: string
}

interface B24NotificationsListProps {
  value: B24Notification[]
  onChange: (notifications: B24Notification[]) => void
  projectName?: string
  branchName?: string
}

export default function B24NotificationsList({
  value,
  onChange,
  projectName,
  branchName,
}: B24NotificationsListProps) {
  const [notifications, setNotifications] = useState<B24Notification[]>(value || [])
  const [testingIndex, setTestingIndex] = useState<number | null>(null)

  useEffect(() => {
    setNotifications(value || [])
  }, [value])

  const handleAdd = () => {
    const next = [...notifications, { type_key: '' }]
    setNotifications(next)
    onChange(next)
  }

  const handleRemove = (index: number) => {
    const next = notifications.filter((_, i) => i !== index)
    setNotifications(next)
    onChange(next)
  }

  const handleChange = (index: number, newValue: string) => {
    const next = [...notifications]
    next[index] = { type_key: newValue }
    setNotifications(next)
    onChange(next)
  }

  const handleTest = async (index: number) => {
    if (!projectName) {
      alert('Project name is required for testing')
      return
    }

    const key = notifications[index]?.type_key.trim()
    if (!key) {
      alert('Please enter type key')
      return
    }

    setTestingIndex(index)
    try {
      if (branchName) {
        await testBranchB24Notification(projectName, branchName, key)
      } else {
        await testProjectB24Notification(projectName, key)
      }
      alert('Test B24 notification sent successfully')
    } catch (error: any) {
      alert('Failed to send test B24 notification: ' + (error.response?.data?.error || error.message))
    } finally {
      setTestingIndex(null)
    }
  }

  return (
    <div className="space-y-2">
      <Label>B24 (Bitrix)</Label>
      <p className="text-xs text-muted-foreground">
        Webhook URL and API key are set in server environment (B24_WEBHOOK_URL, B24_API_KEY).
      </p>
      <div className="space-y-2">
        {notifications.map((notif, index) => (
          <div key={index} className="flex gap-2 items-end">
            <div className="flex-1 space-y-1">
              <Label htmlFor={`b24-type-key-${index}`} className="text-xs">
                type_key
              </Label>
              <Input
                id={`b24-type-key-${index}`}
                type="text"
                value={notif.type_key}
                onChange={(e) => handleChange(index, e.target.value)}
                placeholder="mobile_app"
              />
            </div>
            {projectName && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => handleTest(index)}
                disabled={testingIndex === index}
              >
                {testingIndex === index ? 'Testing...' : 'Test'}
              </Button>
            )}
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={() => handleRemove(index)}
            >
              Remove
            </Button>
          </div>
        ))}
        <Button type="button" variant="outline" onClick={handleAdd}>
          + Add B24 target
        </Button>
      </div>
    </div>
  )
}
