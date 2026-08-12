export default function Modal({ title, subtitle, onClose, children }) {
  return (
    <div className="modal-overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose() }}>
      <div className="modal-card">
        <div className="modal-title">{title}</div>
        {subtitle && <div className="modal-subtitle">{subtitle}</div>}
        {children}
      </div>
    </div>
  )
}
