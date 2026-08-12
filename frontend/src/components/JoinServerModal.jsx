import { useState } from 'react'
import Modal from './Modal'
import { joinServer } from '../lib/api'

export default function JoinServerModal({ onClose, onJoined }) {
  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const server = await joinServer(code.trim())
      onJoined(server)
    } catch (err) {
      setError(err.message || 'Failed to join server')
      setBusy(false)
    }
  }

  return (
    <Modal title="Join a server" subtitle="Enter an invite code to join an existing server" onClose={onClose}>
      {error && <div className="form-error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="invite-code">Invite code</label>
          <input id="invite-code" value={code} onChange={(e) => setCode(e.target.value)} required autoFocus />
        </div>
        <div className="modal-actions">
          <button type="button" className="btn-ghost" onClick={onClose}>Cancel</button>
          <button type="submit" className="btn-accent" disabled={busy || !code.trim()}>
            {busy ? 'Joining…' : 'Join server'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
