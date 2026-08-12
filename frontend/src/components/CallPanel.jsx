import { useEffect, useRef } from 'react'

function VideoTile({ stream, label, isLocal, muted, cameraOff }) {
  const videoRef = useRef(null)

  useEffect(() => {
    if (videoRef.current) videoRef.current.srcObject = stream || null
  }, [stream])

  return (
    <div className="call-tile">
      {stream && !cameraOff ? (
        <video
          ref={videoRef}
          autoPlay
          playsInline
          muted={isLocal}
          className="call-tile__video"
        />
      ) : (
        <div className="call-tile__placeholder">
          <span className="call-tile__initial">{(label || '?').slice(0, 2).toUpperCase()}</span>
        </div>
      )}
      <div className="call-tile__footer">
        <span className="call-tile__name">{label}{isLocal ? ' (you)' : ''}</span>
        {muted && <span className="call-tile__muted" title="Muted">🔇</span>}
      </div>
    </div>
  )
}

export default function CallPanel({ username, localStream, participants, muted, cameraOff, onToggleMute, onToggleCamera, onLeave }) {
  const tiles = [
    { key: 'local', stream: localStream, label: username, isLocal: true, muted, cameraOff },
    ...Array.from(participants.entries()).map(([userID, p]) => ({
      key: userID,
      stream: p.stream,
      label: p.username || 'unknown',
      isLocal: false,
      muted: p.muted,
      cameraOff: p.cameraOff,
    })),
  ]

  return (
    <div className="call-panel">
      <div className="call-grid">
        {tiles.map((t) => (
          <VideoTile
            key={t.key}
            stream={t.stream}
            label={t.label}
            isLocal={t.isLocal}
            muted={t.muted}
            cameraOff={t.cameraOff}
          />
        ))}
      </div>
      <div className="call-controls">
        <button
          className={`call-control-btn ${muted ? 'call-control-btn--active' : ''}`}
          onClick={onToggleMute}
          title={muted ? 'Unmute' : 'Mute'}
        >
          {muted ? '🔇' : '🎙️'}
        </button>
        <button
          className={`call-control-btn ${cameraOff ? 'call-control-btn--active' : ''}`}
          onClick={onToggleCamera}
          title={cameraOff ? 'Turn camera on' : 'Turn camera off'}
        >
          {cameraOff ? '📷' : '🎥'}
        </button>
        <button className="call-control-btn call-control-btn--leave" onClick={onLeave} title="Leave call">
          ⏹ Leave
        </button>
      </div>
    </div>
  )
}