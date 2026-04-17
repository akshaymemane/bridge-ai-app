import { Badge, StatusDot } from './ui/Badge'
import { useApp } from '../context/AppContext'
import type { WsStatus } from '../types'
import { Wifi, WifiOff, Loader2 } from 'lucide-react'
import { formatToolName } from '../lib/utils'

function WsStatusBadge({ status }: { status: WsStatus }) {
  if (status === 'connected') {
    return (
      <Badge variant="online">
        <Wifi size={10} />
        Connected
      </Badge>
    )
  }
  if (status === 'connecting') {
    return (
      <Badge variant="connecting">
        <Loader2 size={10} className="animate-spin" />
        Connecting
      </Badge>
    )
  }
  return (
    <Badge variant="error">
      <WifiOff size={10} />
      Disconnected
    </Badge>
  )
}

export function ChatHeader() {
  const { devices, selectedDeviceId, activeTool, selectTool, wsStatus } = useApp()

  const device = devices.find((d) => d.device_id === selectedDeviceId)
  const availableTools = [...(device?.tools ?? [])].sort((a, b) => {
    const priorities = ['codex', 'claude', 'bridge', 'openclaw', 'ollama']
    const aIndex = priorities.indexOf(a)
    const bIndex = priorities.indexOf(b)
    if (aIndex === -1 && bIndex === -1) return a.localeCompare(b)
    if (aIndex === -1) return 1
    if (bIndex === -1) return -1
    return aIndex - bIndex
  })

  return (
    <header className="flex items-center justify-between px-6 py-3 border-b border-surface-5 bg-surface-2 h-14 shrink-0">
      <div className="flex items-center gap-3">
        {device ? (
          <>
            <div className="w-8 h-8 rounded-lg bg-surface-4 flex items-center justify-center text-xs font-bold text-gray-400 uppercase">
              {device.name.slice(0, 2)}
            </div>
            <div>
              <div className="text-sm font-semibold text-gray-200 leading-tight">
                {device.name}
              </div>
              <div className="flex items-center gap-1.5 flex-wrap">
                <StatusDot status={device.status} />
                <span className="text-[10px] text-gray-500 capitalize">{device.status.replace('_', ' ')}</span>
                {device.tools?.map((tool) => (
                  <Badge
                    key={tool}
                    variant={tool === activeTool ? 'online' : 'offline'}
                    className={tool === activeTool ? 'text-[9px]' : 'border-surface-4 bg-surface-4/70 text-[9px] text-gray-300'}
                  >
                    {formatToolName(tool)}
                  </Badge>
                ))}
              </div>
            </div>
          </>
        ) : (
          <span className="text-sm text-gray-500">No device selected</span>
        )}
      </div>

      <div className="flex items-center gap-3">
        {device && availableTools.length > 0 && (
          <label className="flex items-center gap-2 text-[10px] uppercase tracking-[0.18em] text-gray-500">
            Tool
            <select
              value={activeTool ?? ''}
              onChange={(event) => selectTool(event.target.value)}
              className="rounded-lg border border-surface-5 bg-surface-3 px-2 py-1 text-[11px] font-medium normal-case tracking-normal text-gray-200 outline-none transition focus:border-accent/60"
            >
              {availableTools.map((tool) => (
                <option key={tool} value={tool}>
                  {formatToolName(tool)}
                </option>
              ))}
            </select>
          </label>
        )}
        <WsStatusBadge status={wsStatus} />
      </div>
    </header>
  )
}
