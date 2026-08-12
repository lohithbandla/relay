import { useCallback, useEffect, useState } from 'react'
import { useAuth } from './context/AuthContext'
import AuthScreen from './components/AuthScreen'
import StatusBar from './components/StatusBar'
import ServerRail from './components/ServerRail'
import ChannelSidebar from './components/ChannelSidebar'
import ChatView from './components/ChatView'
import MemberPanel from './components/MemberPanel'
import CreateServerModal from './components/CreateServerModal'
import JoinServerModal from './components/JoinServerModal'
import CreateChannelModal from './components/CreateChannelModal'
import { getMyServers, getChannels, getPresence } from './lib/api'

export default function App() {
  const { user } = useAuth()

  const [servers, setServers] = useState([])
  const [activeServerID, setActiveServerID] = useState(null)
  const [channels, setChannels] = useState([])
  const [activeChannelID, setActiveChannelID] = useState(null)
  const [presence, setPresence] = useState(null)
  const [wsStatus, setWsStatus] = useState('idle')

  const [showCreateServer, setShowCreateServer] = useState(false)
  const [showJoinServer, setShowJoinServer] = useState(false)
  const [showCreateChannel, setShowCreateChannel] = useState(false)

  // Load the user's servers on login.
  useEffect(() => {
    if (!user) return
    getMyServers()
      .then((list) => {
        setServers(list || [])
        if (list?.length) setActiveServerID(list[0].id)
      })
      .catch(() => {})
  }, [user])

  // Load channels + presence whenever the active server changes.
  useEffect(() => {
    if (!activeServerID) {
      setChannels([])
      setActiveChannelID(null)
      setPresence(null)
      return
    }
    getChannels(activeServerID)
      .then((list) => {
        setChannels(list || [])
        setActiveChannelID((prev) => (list?.some((c) => c.id === prev) ? prev : list?.[0]?.id ?? null))
      })
      .catch(() => {})
    getPresence(activeServerID).then(setPresence).catch(() => {})
  }, [activeServerID])

  // Poll presence lightly so the online list stays fresh.
  useEffect(() => {
    if (!activeServerID) return
    const interval = setInterval(() => {
      getPresence(activeServerID).then(setPresence).catch(() => {})
    }, 15000)
    return () => clearInterval(interval)
  }, [activeServerID])

  const handleServerCreated = useCallback((server) => {
    setServers((prev) => [...prev, server])
    setActiveServerID(server.id)
    setShowCreateServer(false)
  }, [])

  const handleServerJoined = useCallback((server) => {
    setServers((prev) => (prev.some((s) => s.id === server.id) ? prev : [...prev, server]))
    setActiveServerID(server.id)
    setShowJoinServer(false)
  }, [])

  const handleChannelCreated = useCallback((channel) => {
    setChannels((prev) => [...prev, channel])
    setActiveChannelID(channel.id)
    setShowCreateChannel(false)
  }, [])

  if (!user) return <AuthScreen />

  const activeServer = servers.find((s) => s.id === activeServerID) || null
  const activeChannel = channels.find((c) => c.id === activeChannelID) || null

  return (
    <div className="app-shell">
      <StatusBar
        wsStatus={wsStatus}
        server={activeServer}
        channel={activeChannel}
        onlineCount={presence?.online_count}
      />
      <div className="app-body">
        <ServerRail
          servers={servers}
          activeServerID={activeServerID}
          onSelect={setActiveServerID}
          onCreate={() => setShowCreateServer(true)}
          onJoin={() => setShowJoinServer(true)}
        />
        <ChannelSidebar
          server={activeServer}
          channels={channels}
          activeChannelID={activeChannelID}
          onSelect={setActiveChannelID}
          onCreateChannel={() => setShowCreateChannel(true)}
        />
        <ChatView channel={activeChannel} onStatusChange={setWsStatus} />
        {activeServer && <MemberPanel presence={presence} />}
      </div>

      {showCreateServer && (
        <CreateServerModal onClose={() => setShowCreateServer(false)} onCreated={handleServerCreated} />
      )}
      {showJoinServer && (
        <JoinServerModal onClose={() => setShowJoinServer(false)} onJoined={handleServerJoined} />
      )}
      {showCreateChannel && activeServer && (
        <CreateChannelModal
          serverID={activeServer.id}
          onClose={() => setShowCreateChannel(false)}
          onCreated={handleChannelCreated}
        />
      )}
    </div>
  )
}
