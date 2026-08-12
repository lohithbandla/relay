import { useCallback, useEffect, useRef, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { getMessages } from '../lib/api'
import { useChannelSocket } from '../hooks/useChannelSocket'
import { useCall } from '../hooks/useCall'
import CallPanel from './CallPanel'

function formatTime(iso) {
  try {
    return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}

const TYPING_TIMEOUT = 4000

export default function ChatView({ channel, onStatusChange }) {
  const { user } = useAuth()
  const [messages, setMessages] = useState([])
  const [systemLog, setSystemLog] = useState([])
  const [loading, setLoading] = useState(true)
  const [draft, setDraft] = useState('')
  const [typingUsers, setTypingUsers] = useState({}) // username -> timeoutId
  const listRef = useRef(null)
  const typingStateRef = useRef(false)
  const typingStopTimerRef = useRef(null)
  // Call signaling events are handled by useCall, which is constructed
  // after `send` exists below — a ref sidesteps that ordering.
  const callSignalRef = useRef(null)

  const onEvent = useCallback((type, payload) => {
    if (type.startsWith('call_')) {
      callSignalRef.current?.(type, payload)
      return
    }
    switch (type) {
      case 'new_message':
        setMessages((prev) => (prev.some((m) => m.id === payload.id) ? prev : [...prev, payload]))
        break
      case 'user_joined':
        setSystemLog((prev) => [...prev, { id: `j-${Date.now()}-${payload.user_id}`, kind: 'join', text: `${payload.username} connected`, ts: Date.now() }])
        break
      case 'user_left':
        setSystemLog((prev) => [...prev, { id: `l-${Date.now()}-${payload.user_id}`, kind: 'leave', text: `${payload.username} disconnected`, ts: Date.now() }])
        break
      case 'typing_start':
        setTypingUsers((prev) => ({ ...prev, [payload.username]: Date.now() }))
        break
      case 'typing_stop':
        setTypingUsers((prev) => {
          const next = { ...prev }
          delete next[payload.username]
          return next
        })
        break
      default:
        break
    }
  }, [])

  const { status, send } = useChannelSocket(channel?.id, onEvent)
  const call = useCall(channel?.id, send, status)

  useEffect(() => {
    callSignalRef.current = call.handleSignalEvent
  }, [call.handleSignalEvent])

  useEffect(() => {
    onStatusChange?.(status)
  }, [status, onStatusChange])

  // Load history whenever the channel changes.
  useEffect(() => {
    if (!channel?.id) return
    let cancelled = false
    setLoading(true)
    setMessages([])
    setSystemLog([])
    getMessages(channel.id, 50, 0)
      .then((data) => {
        if (cancelled) return
        setMessages([...(data.messages || [])].reverse())
      })
      .catch(() => {})
      .finally(() => !cancelled && setLoading(false))
    return () => { cancelled = true }
  }, [channel?.id])

  // Auto-expire stale typing indicators.
  useEffect(() => {
    const interval = setInterval(() => {
      setTypingUsers((prev) => {
        const now = Date.now()
        const next = {}
        for (const [name, ts] of Object.entries(prev)) {
          if (now - ts < TYPING_TIMEOUT) next[name] = ts
        }
        return next
      })
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  // Scroll to bottom on new content.
  useEffect(() => {
    listRef.current?.scrollTo({ top: listRef.current.scrollHeight })
  }, [messages, systemLog])

  function handleDraftChange(e) {
    setDraft(e.target.value)
    if (!typingStateRef.current) {
      typingStateRef.current = true
      send('typing_start', {})
    }
    clearTimeout(typingStopTimerRef.current)
    typingStopTimerRef.current = setTimeout(() => {
      typingStateRef.current = false
      send('typing_stop', {})
    }, 2000)
  }

  function handleSend(e) {
    e.preventDefault()
    const content = draft.trim()
    if (!content) return
    send('message', { content })
    setDraft('')
    clearTimeout(typingStopTimerRef.current)
    if (typingStateRef.current) {
      typingStateRef.current = false
      send('typing_stop', {})
    }
  }

  if (!channel) {
    return (
      <div className="chat-view">
        <div className="empty-state">// select or create a channel to start relaying messages</div>
      </div>
    )
  }

  const feed = [
    ...messages.map((m) => ({ ...m, kind: 'message', ts: new Date(m.created_at).getTime() })),
    ...systemLog,
  ].sort((a, b) => a.ts - b.ts)

  const typingNames = Object.keys(typingUsers)

  return (
    <div className="chat-view">
      <div className="chat-header">
        <span className="chat-header__hash">{channel.type === 'voice' ? '🔊' : '#'}</span>
        <span className="chat-header__name">{channel.name}</span>
        {channel.topic && <span className="chat-header__topic">{channel.topic}</span>}
        <button
          className={`call-header-btn ${call.inCall ? 'call-header-btn--active' : ''}`}
          onClick={call.inCall ? call.leaveCall : call.joinCall}
          disabled={status !== 'open'}
        >
          {call.inCall ? '⏹ Leave call' : '📹 Start call'}
        </button>
      </div>

      {call.mediaError && (
        <div className="call-error">// couldn't access camera/mic: {call.mediaError}</div>
      )}

      {call.inCall && (
        <CallPanel
          username={user?.username}
          localStream={call.localStream}
          participants={call.participants}
          muted={call.muted}
          cameraOff={call.cameraOff}
          onToggleMute={call.toggleMute}
          onToggleCamera={call.toggleCamera}
          onLeave={call.leaveCall}
        />
      )}

      <div className="message-list" ref={listRef}>
        {loading && <div className="center-loading">// loading history</div>}
        {!loading && feed.length === 0 && (
          <div className="empty-state">// #{channel.name} is quiet. Send the first message.</div>
        )}
        {feed.map((item) =>
          item.kind === 'message' ? (
            <div className="message-row" key={item.id}>
              <div className="message-avatar">{(item.sender?.username || '?').slice(0, 2).toUpperCase()}</div>
              <div className="message-body">
                <div className="message-meta">
                  <span className="message-author">{item.sender?.username || 'unknown'}</span>
                  <span className="message-time">{formatTime(item.created_at)}</span>
                </div>
                <div className="message-content">{item.content}</div>
              </div>
            </div>
          ) : (
            <div className={`system-line system-line--${item.kind}`} key={item.id}>
              {item.kind === 'join' ? '→' : '←'} {item.text}
            </div>
          )
        )}
      </div>

      <div className="typing-strip">
        {typingNames.length > 0 &&
          `${typingNames.join(', ')} ${typingNames.length === 1 ? 'is' : 'are'} typing…`}
      </div>

      <form className="composer" onSubmit={handleSend}>
        <div className="composer__inner">
          <input
            className="composer__input"
            placeholder={status === 'open' ? `Message #${channel.name}` : 'Reconnecting…'}
            value={draft}
            onChange={handleDraftChange}
            disabled={status !== 'open'}
          />
          <button className="composer__send" type="submit" disabled={status !== 'open' || !draft.trim()}>
            Send
          </button>
        </div>
      </form>
    </div>
  )
}