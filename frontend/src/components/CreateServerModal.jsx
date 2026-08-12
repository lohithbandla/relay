import { useState } from 'react'
import Modal from './Modal'
import { createServer } from '../lib/api'

export default function CreateServerModal({ onClose, onCreated }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const server = await createServer(name.trim(), description.trim())
      onCreated(server)
    } catch (err) {
      setError(err.message || 'Failed to create server')
      setBusy(false)
    }
  }

  return (
    <Modal title="New server" subtitle="Spin up a new space for your people" onClose={onClose}>
      {error && <div className="form-error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="server-name">Name</label>
          <input id="server-name" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
        </div>
        <div className="field">
          <label htmlFor="server-desc">Description (optional)</label>
          <input id="server-desc" value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        <div className="modal-actions">
          <button type="button" className="btn-ghost" onClick={onClose}>Cancel</button>
          <button type="submit" className="btn-accent" disabled={busy || !name.trim()}>
            {busy ? 'Creating…' : 'Create server'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
