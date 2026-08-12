import { useCallback, useEffect, useRef, useState } from 'react'

// Public STUN server — enough to discover a host's reflexive address for
// testing across NATs. No TURN relay configured, so calls between peers on
// strict/symmetric NATs may fail to connect; that's a real-world limitation
// of this project, not a bug.
const ICE_SERVERS = [{ urls: 'stun:stun.l.google.com:19302' }]

// Manages an in-call WebRTC mesh for one channel: one RTCPeerConnection per
// other participant, all signaled over the same channel WebSocket that
// text chat already uses (the `send` function passed in). The backend
// never touches media, only offer/answer/ICE routing.
//
// Convention that avoids SDP glare: when you join, the server tells you
// who's already there via `call_participants`, and *you* send offers to
// each of them. Everyone else just waits for your offer and answers it.
export function useCall(channelID, send, wsStatus) {
  const [inCall, setInCall] = useState(false)
  const [localStream, setLocalStream] = useState(null)
  const [participants, setParticipants] = useState(new Map()) // userID -> { username, stream, muted, cameraOff }
  const [muted, setMuted] = useState(false)
  const [cameraOff, setCameraOff] = useState(false)
  const [mediaError, setMediaError] = useState(null)

  const localStreamRef = useRef(null)
  const sendRef = useRef(send)
  sendRef.current = send
  const peerConnectionsRef = useRef(new Map()) // userID -> RTCPeerConnection
  const pendingCandidatesRef = useRef(new Map()) // userID -> candidate[] (arrived before remote description was set)

  const createPeerConnection = useCallback((peerUserID) => {
    const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS })

    localStreamRef.current?.getTracks().forEach((track) => {
      pc.addTrack(track, localStreamRef.current)
    })

    pc.onicecandidate = (e) => {
      if (e.candidate) {
        sendRef.current('call_ice_candidate', {
          target_user_id: peerUserID,
          candidate: e.candidate,
        })
      }
    }

    pc.ontrack = (e) => {
      setParticipants((prev) => {
        const next = new Map(prev)
        const existing = next.get(peerUserID) || {}
        next.set(peerUserID, { ...existing, stream: e.streams[0] })
        return next
      })
    }

    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'failed') {
        pc.restartIce?.()
      }
    }

    peerConnectionsRef.current.set(peerUserID, pc)
    return pc
  }, [])

  const flushPendingCandidates = useCallback(async (userID, pc) => {
    const queue = pendingCandidatesRef.current.get(userID)
    if (!queue || queue.length === 0) return
    for (const candidate of queue) {
      try {
        await pc.addIceCandidate(candidate)
      } catch {
        // ignore malformed/late candidates
      }
    }
    pendingCandidatesRef.current.delete(userID)
  }, [])

  const teardownPeer = useCallback((userID) => {
    const pc = peerConnectionsRef.current.get(userID)
    pc?.close()
    peerConnectionsRef.current.delete(userID)
    pendingCandidatesRef.current.delete(userID)
    setParticipants((prev) => {
      if (!prev.has(userID)) return prev
      const next = new Map(prev)
      next.delete(userID)
      return next
    })
  }, [])

  const teardownAllPeers = useCallback(() => {
    for (const pc of peerConnectionsRef.current.values()) pc.close()
    peerConnectionsRef.current.clear()
    pendingCandidatesRef.current.clear()
    setParticipants(new Map())
  }, [])

  const joinCall = useCallback(async () => {
    if (inCall) return
    setMediaError(null)
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      localStreamRef.current = stream
      setLocalStream(stream)
      setMuted(false)
      setCameraOff(false)
      setInCall(true)
      sendRef.current('call_join', {})
    } catch (err) {
      setMediaError(err?.message || 'Could not access camera/microphone')
    }
  }, [inCall])

  const leaveCall = useCallback(() => {
    if (!inCall) return
    sendRef.current('call_leave', {})
    teardownAllPeers()
    localStreamRef.current?.getTracks().forEach((t) => t.stop())
    localStreamRef.current = null
    setLocalStream(null)
    setInCall(false)
    setMuted(false)
    setCameraOff(false)
  }, [inCall, teardownAllPeers])

  const toggleMute = useCallback(() => {
    const stream = localStreamRef.current
    if (!stream) return
    const next = !muted
    stream.getAudioTracks().forEach((t) => { t.enabled = !next })
    setMuted(next)
    sendRef.current('call_mute', { muted: next })
  }, [muted])

  const toggleCamera = useCallback(() => {
    const stream = localStreamRef.current
    if (!stream) return
    const next = !cameraOff
    stream.getVideoTracks().forEach((t) => { t.enabled = !next })
    setCameraOff(next)
    sendRef.current('call_camera', { camera_off: next })
  }, [cameraOff])

  // Handles every `call_*` event coming off the channel WebSocket.
  // ChatView delegates to this from its own onEvent switch.
  const handleSignalEvent = useCallback(async (type, payload) => {
    switch (type) {
      case 'call_participants': {
        for (const p of payload.participants || []) {
          setParticipants((prev) => new Map(prev).set(p.user_id, {
            username: p.username, stream: null, muted: p.muted, cameraOff: p.camera_off,
          }))
          const pc = createPeerConnection(p.user_id)
          const offer = await pc.createOffer()
          await pc.setLocalDescription(offer)
          sendRef.current('call_offer', { target_user_id: p.user_id, sdp: offer.sdp })
        }
        break
      }

      case 'call_user_joined': {
        setParticipants((prev) => new Map(prev).set(payload.user_id, {
          username: payload.username, stream: null, muted: false, cameraOff: false,
        }))
        break
      }

      case 'call_user_left': {
        teardownPeer(payload.user_id)
        break
      }

      case 'call_offer': {
        let pc = peerConnectionsRef.current.get(payload.from_user_id)
        if (!pc) pc = createPeerConnection(payload.from_user_id)
        await pc.setRemoteDescription({ type: 'offer', sdp: payload.sdp })
        await flushPendingCandidates(payload.from_user_id, pc)
        const answer = await pc.createAnswer()
        await pc.setLocalDescription(answer)
        sendRef.current('call_answer', { target_user_id: payload.from_user_id, sdp: answer.sdp })
        break
      }

      case 'call_answer': {
        const pc = peerConnectionsRef.current.get(payload.from_user_id)
        if (pc) {
          await pc.setRemoteDescription({ type: 'answer', sdp: payload.sdp })
          await flushPendingCandidates(payload.from_user_id, pc)
        }
        break
      }

      case 'call_ice_candidate': {
        const pc = peerConnectionsRef.current.get(payload.from_user_id)
        if (pc && pc.remoteDescription) {
          try {
            await pc.addIceCandidate(payload.candidate)
          } catch {
            // ignore
          }
        } else {
          const queue = pendingCandidatesRef.current.get(payload.from_user_id) || []
          queue.push(payload.candidate)
          pendingCandidatesRef.current.set(payload.from_user_id, queue)
        }
        break
      }

      case 'call_user_muted': {
        setParticipants((prev) => {
          const existing = prev.get(payload.user_id)
          if (!existing) return prev
          const next = new Map(prev)
          next.set(payload.user_id, { ...existing, muted: payload.muted })
          return next
        })
        break
      }

      case 'call_user_camera': {
        setParticipants((prev) => {
          const existing = prev.get(payload.user_id)
          if (!existing) return prev
          const next = new Map(prev)
          next.set(payload.user_id, { ...existing, cameraOff: payload.camera_off })
          return next
        })
        break
      }

      default:
        break
    }
  }, [createPeerConnection, flushPendingCandidates, teardownPeer])

  // Leave the call if the channel changes or the component unmounts —
  // don't leak a live camera/mic or dangling peer connections.
  useEffect(() => {
    return () => {
      teardownAllPeers()
      localStreamRef.current?.getTracks().forEach((t) => t.stop())
      localStreamRef.current = null
    }
  }, [channelID, teardownAllPeers])

  // If the socket drops, the backend already removes us from the call
  // server-side. Reflect that locally too instead of showing a call UI
  // that's silently disconnected.
  useEffect(() => {
    if (wsStatus !== 'open' && inCall) {
      teardownAllPeers()
      localStreamRef.current?.getTracks().forEach((t) => t.stop())
      localStreamRef.current = null
      setLocalStream(null)
      setInCall(false)
    }
  }, [wsStatus, inCall, teardownAllPeers])

  return {
    inCall,
    localStream,
    participants,
    muted,
    cameraOff,
    mediaError,
    joinCall,
    leaveCall,
    toggleMute,
    toggleCamera,
    handleSignalEvent,
  }
}