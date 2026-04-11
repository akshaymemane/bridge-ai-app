import { Badge, StatusDot } from './ui/Badge'
import { useApp } from '../context/AppContext'
import type { WsStatus } from '../types'
import { Wifi, WifiOff, Loader2 } from 'lucide-react'

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
  const { devices, selectedDeviceId, wsStatus } = useApp()

  const device = devices.find((d) => d.device_id === selectedDeviceId)

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
              <div className="flex items-center gap-1.5">
                <StatusDot status={device.status} />
                <span className="text-[10px] text-gray-500 capitalize">{device.status}</span>
              </div>
            </div>
          </>
        ) : (
          <span className="text-sm text-gray-500">No device selected</span>
        )}
      </div>

      <WsStatusBadge status={wsStatus} />
    </header>
  )
}
