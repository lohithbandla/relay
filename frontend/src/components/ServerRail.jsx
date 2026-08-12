function initials(name) {
  return (name || '?').trim().slice(0, 2).toUpperCase()
}

export default function ServerRail({ servers, activeServerID, onSelect, onCreate, onJoin }) {
  return (
    <div className="server-rail">
      {servers.map((s) => (
        <div
          key={s.id}
          className={`server-icon ${s.id === activeServerID ? 'server-icon--active' : ''}`}
          onClick={() => onSelect(s.id)}
          title={s.name}
        >
          {initials(s.name)}
        </div>
      ))}
      <div className="server-rail__divider" />
      <div className="server-icon server-icon--ghost" onClick={onCreate} title="Create a server">+</div>
      <div className="server-icon server-icon--ghost" onClick={onJoin} title="Join a server">↵</div>
    </div>
  )
}
