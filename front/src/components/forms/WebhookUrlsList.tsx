import { useState, useEffect } from 'react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { testProjectWebhook } from '../../api'
import { testBranchWebhook } from '../../api/branches'

interface WebhookUrlsListProps {
  value: string[]
  onChange: (urls: string[]) => void
  projectName?: string
  branchName?: string
}

export default function WebhookUrlsList({ value, onChange, projectName, branchName }: WebhookUrlsListProps) {
  const [urls, setUrls] = useState<string[]>(value || [])
  const [testingIndex, setTestingIndex] = useState<number | null>(null)

  useEffect(() => {
    setUrls(value || [])
  }, [value])

  const handleAdd = () => {
    const newUrls = [...urls, '']
    setUrls(newUrls)
    onChange(newUrls)
  }

  const handleRemove = (index: number) => {
    const newUrls = urls.filter((_, i) => i !== index)
    setUrls(newUrls)
    onChange(newUrls)
  }

  const handleChange = (index: number, newValue: string) => {
    const newUrls = [...urls]
    newUrls[index] = newValue
    setUrls(newUrls)
    onChange(newUrls)
  }

  const handleTest = async (index: number) => {
    if (!projectName) {
      alert('Project name is required for testing')
      return
    }

    const url = urls[index]
    if (!url.trim()) {
      alert('Please enter webhook URL')
      return
    }

    setTestingIndex(index)
    try {
      if (branchName) {
        await testBranchWebhook(projectName, branchName, url.trim())
      } else {
        await testProjectWebhook(projectName, url.trim())
      }
      alert('Test webhook sent successfully')
    } catch (error: any) {
      alert('Failed to send test webhook: ' + (error.response?.data?.error || error.message))
    } finally {
      setTestingIndex(null)
    }
  }

  return (
    <div className="space-y-2">
      <Label>Webhook URLs</Label>
      <div className="space-y-2">
        {urls.map((url, index) => (
          <div key={index} className="flex gap-2 items-end">
            <div className="flex-1">
              <Input
                type="text"
                value={url}
                onChange={(e) => handleChange(index, e.target.value)}
                placeholder="https://portainer.example.com/webhook"
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
          + Add Webhook
        </Button>
      </div>
    </div>
  )
}

