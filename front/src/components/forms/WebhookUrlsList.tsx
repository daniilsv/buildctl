import { useState, useEffect } from 'react'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

interface WebhookUrlsListProps {
  value: string[]
  onChange: (urls: string[]) => void
}

export default function WebhookUrlsList({ value, onChange }: WebhookUrlsListProps) {
  const [urls, setUrls] = useState<string[]>(value || [])

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

