import { AppProvider } from './context/AppContext'
import { Sidebar } from './components/Sidebar'
import { ChatHeader } from './components/ChatHeader'
import { ChatThread } from './components/ChatThread'
import { MessageInput } from './components/MessageInput'
import { ReconnectBanner } from './components/ReconnectBanner'
import { LoginScreen } from './components/LoginScreen'
import { useApp } from './context/AppContext'
import { Loader2 } from 'lucide-react'

function ChatLayout() {
  const { authLoading, isAuthenticated } = useApp()

  if (authLoading) {
    return (
      <div className="flex h-full items-center justify-center bg-surface-0 text-gray-400">
        <div className="flex items-center gap-3 rounded-2xl border border-surface-5 bg-surface-2 px-5 py-4">
          <Loader2 size={16} className="animate-spin" />
          <span className="text-sm">Loading BridgeAI…</span>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <LoginScreen />
  }

  return (
    <div className="flex h-full w-full overflow-hidden bg-surface-0">
      {/* Sidebar */}
      <Sidebar />

      {/* Main chat area */}
      <div className="flex flex-col flex-1 min-w-0 h-full">
        {/* Reconnect banner (full width, above header) */}
        <ReconnectBanner />

        {/* Device header */}
        <ChatHeader />

        {/* Messages */}
        <ChatThread />

        {/* Input bar */}
        <MessageInput />
      </div>
    </div>
  )
}

export default function App() {
  return (
    <AppProvider>
      <ChatLayout />
    </AppProvider>
  )
}
