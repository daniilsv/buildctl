import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getSSHKeys } from '../../api/ssh_keys'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Select } from '../ui/select'

export interface SSHAction {
  host: string
  port: number
  username: string
  ssh_key_id: string
  command: string
}

interface SSHActionsListProps {
  value: SSHAction[]
  onChange: (actions: SSHAction[]) => void
}

function normalizeAction(a: any): SSHAction {
  return {
    host: String(a?.host || ''),
    port: typeof a?.port === 'number' ? a.port : parseInt(String(a?.port || '22'), 10) || 22,
    username: String(a?.username || ''),
    ssh_key_id: String(a?.ssh_key_id || ''),
    command: String(a?.command || ''),
  }
}

export default function SSHActionsList({ value, onChange }: SSHActionsListProps) {
  const [actions, setActions] = useState<SSHAction[]>(() =>
    (value || []).map(normalizeAction)
  )

  const { data: sshKeys = [] } = useQuery({
    queryKey: ['ssh-keys'],
    queryFn: getSSHKeys,
  })

  useEffect(() => {
    setActions((value || []).map(normalizeAction))
  }, [value])

  const handleAdd = () => {
    const newActions = [...actions, { host: '', port: 22, username: '', ssh_key_id: '', command: '' }]
    setActions(newActions)
    onChange(newActions)
  }

  const handleRemove = (index: number) => {
    const newActions = actions.filter((_, i) => i !== index)
    setActions(newActions)
    onChange(newActions)
  }

  const handleChange = (index: number, field: keyof SSHAction, newValue: string | number) => {
    const newActions = [...actions]
    newActions[index] = { ...newActions[index], [field]: newValue }
    setActions(newActions)
    onChange(newActions)
  }

  return (
    <div className="space-y-2">
      <Label>SSH Deploy Actions</Label>
      <p className="text-sm text-muted-foreground">
        Run commands on remote servers after successful build. Add SSH keys in Settings first.
      </p>
      <div className="space-y-4">
        {actions.map((action, index) => (
          <div key={index} className="rounded-lg border p-4 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>Host</Label>
                <Input
                  type="text"
                  value={action.host}
                  onChange={(e) => handleChange(index, 'host', e.target.value)}
                  placeholder="deploy.example.com"
                />
              </div>
              <div className="space-y-2">
                <Label>Port</Label>
                <Input
                  type="number"
                  value={action.port}
                  onChange={(e) => handleChange(index, 'port', parseInt(e.target.value) || 22)}
                  placeholder="22"
                />
              </div>
              <div className="space-y-2">
                <Label>Username</Label>
                <Input
                  type="text"
                  value={action.username}
                  onChange={(e) => handleChange(index, 'username', e.target.value)}
                  placeholder="deploy"
                />
              </div>
              <div className="space-y-2">
                <Label>SSH Key</Label>
                <Select
                  value={action.ssh_key_id}
                  onChange={(e) => handleChange(index, 'ssh_key_id', e.target.value)}
                >
                  <option value="">Select key...</option>
                  {sshKeys.map((key) => (
                    <option key={key.id} value={key.id}>
                      {key.name} ({key.fingerprint})
                    </option>
                  ))}
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>Command</Label>
              <textarea
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
                value={action.command}
                onChange={(e) => handleChange(index, 'command', e.target.value)}
                placeholder="cd /app && docker-compose pull && docker-compose up -d"
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
          + Add SSH Action
        </Button>
      </div>
    </div>
  )
}
