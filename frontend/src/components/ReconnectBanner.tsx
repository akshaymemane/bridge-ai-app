import { useApp } from '../context/AppContext'
import { WifiOff, Loader2 } from 'lucide-react'

export function ReconnectBanner() {
  const { wsStatus } = useApp()

  if (wsStatus === 'connected') return null

  const isConnecting = wsStatus === 'connecting'

  return (
    <div
      role="alert"
      className="flex items-center justify-center gap-2 px-4 py-2 bg-yellow-950/80 border-b border-yellow-700/40 text-yellow-400 text-xs w-full"
    >
      {isConnecting ? (
        <>
          <Loader2 size={12} className="animate-spin" />
          <span>Connecting to gateway…</span>
        </>
      ) : (
        <>
          <WifiOff size={12} />
          <span>
            Gateway disconnected — attempting to reconnect automatically. Check that the gateway is
            running on <code className="font-mono-chat text-yellow-300">{window.location.host}</code>.
          </span>
        </>
      )}
    </div>
  )
}
