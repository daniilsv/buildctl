import { useState, useEffect } from 'react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { testProjectTelegramNotification, testBranchTelegramNotification } from '../../api'

interface TelegramNotification {
  chat_id: string
  thread_id: string
}

interface TelegramNotificationsListProps {
  value: TelegramNotification[]
  onChange: (notifications: TelegramNotification[]) => void
  projectName?: string
  branchName?: string
}

export default function TelegramNotificationsList({
  value,
  onChange,
  projectName,
  branchName,
}: TelegramNotificationsListProps) {
  const [notifications, setNotifications] = useState<TelegramNotification[]>(value || [])
  const [testingIndex, setTestingIndex] = useState<number | null>(null)

  useEffect(() => {
    setNotifications(value || [])
  }, [value])

  const handleAdd = () => {
    const newNotifications = [...notifications, { chat_id: '', thread_id: '' }]
    setNotifications(newNotifications)
    onChange(newNotifications)
  }

  const handleRemove = (index: number) => {
    const newNotifications = notifications.filter((_, i) => i !== index)
    setNotifications(newNotifications)
    onChange(newNotifications)
  }

  const handleChange = (index: number, field: 'chat_id' | 'thread_id', newValue: string) => {
    const newNotifications = [...notifications]
    newNotifications[index] = {
      ...newNotifications[index],
      [field]: newValue,
    }
    setNotifications(newNotifications)
    onChange(newNotifications)
  }

  const handleTest = async (index: number) => {
    if (!projectName) {
      alert('Project name is required for testing')
      return
    }

    const notif = notifications[index]
    if (!notif.chat_id.trim()) {
      alert('Please enter Chat ID')
      return
    }

    setTestingIndex(index)
    try {
      if (branchName) {
        await testBranchTelegramNotification(projectName, branchName, notif.chat_id.trim(), notif.thread_id.trim() || undefined)
      } else {
        await testProjectTelegramNotification(projectName, notif.chat_id.trim(), notif.thread_id.trim() || undefined)
      }
      alert('Test notification sent successfully')
    } catch (error: any) {
      alert('Failed to send test notification: ' + (error.response?.data?.error || error.message))
    } finally {
      setTestingIndex(null)
    }
  }

  return (
    <div className="space-y-2">
      <Label>Telegram Notifications</Label>
      <div className="space-y-2">
        {notifications.map((notif, index) => (
          <div key={index} className="flex gap-2 items-end">
            <div className="flex-1 space-y-1">
              <Label htmlFor={`chat-id-${index}`} className="text-xs">
                Chat ID
              </Label>
              <Input
                id={`chat-id-${index}`}
                type="text"
                value={notif.chat_id}
                onChange={(e) => handleChange(index, 'chat_id', e.target.value)}
                placeholder="-4859320492"
              />
            </div>
            <div className="flex-1 space-y-1">
              <Label htmlFor={`thread-id-${index}`} className="text-xs">
                Thread ID (optional)
              </Label>
              <Input
                id={`thread-id-${index}`}
                type="text"
                value={notif.thread_id}
                onChange={(e) => handleChange(index, 'thread_id', e.target.value)}
                placeholder="123"
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
          + Add Notification
        </Button>
      </div>
    </div>
  )
}

