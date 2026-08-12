import { useEffect, useRef, useState, useCallback } from 'react'
import { wsUrl } from '../lib/api'

// Manages one WebSocket connection per active channel.
// Mirrors the backend's event envelope: { type, payload }.
export function useChannelSocket(channelID, onEvent) {
  const [status, setStatus] = useState('idle') // idle | connecting | open | closed | error
  const socketRef = useRef(null)
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!channelID) return

    let cancelled = false
    let retryTimer = null
    let retryDelay = 1000

    function connect() {
      if (cancelled) return
      setStatus('connecting')
      const socket = new WebSocket(wsUrl(channelID))
      socketRef.current = socket

      socket.onopen = () => {
        if (cancelled) return
        retryDelay = 1000
        setStatus('open')
      }

      socket.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data)
          onEventRef.current?.(parsed.type, parsed.payload)
        } catch {
          // ignore malformed frames
        }
      }

      socket.onclose = () => {
        if (cancelled) return
        setStatus('closed')
        retryTimer = setTimeout(connect, retryDelay)
        retryDelay = Math.min(retryDelay * 2, 10000)
      }

      socket.onerror = () => {
        setStatus('error')
      }
    }

    connect()

    return () => {
      cancelled = true
      clearTimeout(retryTimer)
      socketRef.current?.close()
      socketRef.current = null
      setStatus('idle')
    }
  }, [channelID])

  const send = useCallback((type, payload) => {
    const socket = socketRef.current
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type, payload }))
    }
  }, [])

  return { status, send }
}
