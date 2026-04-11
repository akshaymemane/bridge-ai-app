import { AppProvider } from './context/AppContext'
import { Sidebar } from './components/Sidebar'
import { ChatHeader } from './components/ChatHeader'
import { ChatThread } from './components/ChatThread'
import { MessageInput } from './components/MessageInput'
import { ReconnectBanner } from './components/ReconnectBanner'

function ChatLayout() {
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
