import { useAuth } from '../context/AuthContext'

const STATUS_LABEL = {
  idle: 'idle',
  connecting: 'connecting',
  open: 'connected',
  closed: 'reconnecting',
  error: 'error',
}

export default function StatusBar({ wsStatus, server, channel, onlineCount }) {
  const { user, logout } = useAuth()

  return (
    <div className="status-bar">
      <div className="status-bar__brand">⚡ RELAY</div>
      <div className="status-bar__divider">/</div>
      <div className="status-bar__item">
        <span className={`status-dot status-dot--${wsStatus}`} />
        ws: {STATUS_LABEL[wsStatus] || wsStatus}
      </div>
      {server && (
        <div className="status-bar__item">
          {server.name}{channel ? ` / #${channel.name}` : ''}
        </div>
      )}
      {typeof onlineCount === 'number' && (
        <div className="status-bar__item">online: {onlineCount}</div>
      )}
      <div className="status-bar__spacer" />
      <div className="status-bar__user">{user?.username}</div>
      <button className="status-bar__logout" onClick={logout}>logout</button>
    </div>
  )
}
