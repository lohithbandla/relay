import { useAuth } from '../context/AuthContext'

function initials(name) {
  return (name || '?').trim().slice(0, 2).toUpperCase()
}

export default function ChannelSidebar({ server, channels, activeChannelID, onSelect, onCreateChannel }) {
  const { user } = useAuth()

  if (!server) {
    return (
      <div className="channel-sidebar">
        <div className="center-loading" style={{ margin: 'auto' }}>// no server selected</div>
      </div>
    )
  }

  return (
    <div className="channel-sidebar">
      <div className="channel-sidebar__header">
        <div className="channel-sidebar__server-name">{server.name}</div>
        <div className="channel-sidebar__invite">invite: {server.invite_code}</div>
      </div>

      <div className="channel-sidebar__section">
        <div className="channel-sidebar__label">
          <span>Channels</span>
          <button className="channel-sidebar__add" onClick={onCreateChannel} title="Create channel">+</button>
        </div>
        {channels.length === 0 && (
          <div className="empty-state" style={{ margin: '20px 4px' }}>no channels yet</div>
        )}
        {channels.map((c) => (
          <div
            key={c.id}
            className={`channel-row ${c.id === activeChannelID ? 'channel-row--active' : ''}`}
            onClick={() => onSelect(c.id)}
          >
            <span className="channel-row__hash">{c.type === 'voice' ? '🔊' : '#'}</span>
            <span>{c.name}</span>
          </div>
        ))}
      </div>

      <div className="channel-sidebar__footer">
        <div className="channel-sidebar__avatar">{initials(user?.username)}</div>
        <div>{user?.username}</div>
      </div>
    </div>
  )
}
