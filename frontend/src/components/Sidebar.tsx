import { ScrollArea } from './ui/ScrollArea'
import { StatusDot } from './ui/Badge'
import { useApp } from '../context/AppContext'
import { cn } from '../lib/utils'
import { ServerCrash, Loader2 } from 'lucide-react'

export function Sidebar() {
  const { devices, devicesLoading, devicesError, selectedDeviceId, selectDevice } = useApp()

  return (
    <aside className="flex flex-col w-64 shrink-0 border-r border-surface-5 bg-surface-1 h-full">
      {/* Logo / App name */}
      <div className="flex items-center gap-2.5 px-4 py-4 border-b border-surface-5">
        <div className="w-7 h-7 rounded-lg bg-accent flex items-center justify-center text-white text-xs font-bold select-none">
          B
        </div>
        <span className="text-sm font-semibold text-gray-200 tracking-tight">BridgeAIChat</span>
      </div>

      {/* Devices section label */}
      <div className="px-4 pt-4 pb-1.5">
        <span className="text-[10px] font-semibold uppercase tracking-widest text-gray-500">
          Devices
        </span>
      </div>

      {/* Device list */}
      <ScrollArea className="flex-1 px-2 pb-4">
        {devicesLoading && (
          <div className="flex items-center gap-2 px-3 py-3 text-gray-500 text-sm">
            <Loader2 size={14} className="animate-spin" />
            <span>Loading devices…</span>
          </div>
        )}

        {devicesError && !devicesLoading && (
          <div className="flex items-start gap-2 px-3 py-3 text-red-400 text-xs">
            <ServerCrash size={14} className="mt-0.5 shrink-0" />
            <span>{devicesError}</span>
          </div>
        )}

        {!devicesLoading && !devicesError && devices.length === 0 && (
          <div className="px-3 py-3 text-gray-500 text-xs">
            No devices registered.
          </div>
        )}

        {devices.map((device) => {
          const isSelected = device.device_id === selectedDeviceId
          return (
            <button
              key={device.device_id}
              onClick={() => selectDevice(device.device_id)}
              className={cn(
                'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors group',
                isSelected
                  ? 'bg-accent/15 text-gray-100'
                  : 'text-gray-400 hover:bg-surface-4 hover:text-gray-200'
              )}
            >
              {/* Device icon */}
              <div
                className={cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center text-xs font-bold shrink-0 uppercase',
                  isSelected
                    ? 'bg-accent text-white'
                    : 'bg-surface-4 text-gray-500 group-hover:bg-surface-5'
                )}
              >
                {device.name.slice(0, 2)}
              </div>

              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate leading-tight">
                  {device.name}
                </div>
                <div className="flex items-center gap-1.5 mt-0.5">
                  <StatusDot status={device.status} />
                  <span className="text-[10px] text-gray-500 capitalize">{device.status}</span>
                </div>
              </div>
            </button>
          )
        })}
      </ScrollArea>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-surface-5">
        <p className="text-[10px] text-gray-600">v1.0.0-alpha</p>
      </div>
    </aside>
  )
}
