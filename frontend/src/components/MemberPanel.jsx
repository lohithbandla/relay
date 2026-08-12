import { useAuth } from '../context/AuthContext'

export default function MemberPanel({ presence }) {
  const { user } = useAuth()

  if (!presence) {
    return (
      <div className="member-panel">
        <div className="center-loading">// loading presence</div>
      </div>
    )
  }

  const entries = Object.entries(presence.presence || {}).sort((a, b) => Number(b[1]) - Number(a[1]))

  return (
    <div className="member-panel">
      <div className="member-panel__label">Online — {presence.online_count ?? 0}</div>
      {entries.length === 0 && <div className="empty-state" style={{ margin: '16px 0' }}>no members</div>}
      {entries.map(([id, online]) => (
        <div key={id} className={`member-row ${online ? 'member-row--online' : ''}`}>
          <span className={`member-dot ${online ? 'member-dot--online' : ''}`} />
          <span>{id === user?.id ? `${user.username} (you)` : `${id.slice(0, 8)}…`}</span>
        </div>
      ))}
    </div>
  )
}
