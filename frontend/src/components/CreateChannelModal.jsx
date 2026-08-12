import { useState } from 'react'
import Modal from './Modal'
import { createChannel } from '../lib/api'

export default function CreateChannelModal({ serverID, onClose, onCreated }) {
  const [name, setName] = useState('')
  const [topic, setTopic] = useState('')
  const [type, setType] = useState('text')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const channel = await createChannel(serverID, name.trim(), topic.trim(), type)
      onCreated(channel)
    } catch (err) {
      setError(err.message || 'Failed to create channel')
      setBusy(false)
    }
  }

  return (
    <Modal title="New channel" subtitle="Add a channel to this server" onClose={onClose}>
      {error && <div className="form-error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="channel-name">Name</label>
          <input id="channel-name" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
        </div>
        <div className="field">
          <label htmlFor="channel-topic">Topic (optional)</label>
          <input id="channel-topic" value={topic} onChange={(e) => setTopic(e.target.value)} />
        </div>
        <div className="field">
          <label htmlFor="channel-type">Type</label>
          <input id="channel-type" value={type} onChange={(e) => setType(e.target.value)} list="channel-types" />
          <datalist id="channel-types">
            <option value="text" />
            <option value="voice" />
          </datalist>
        </div>
        <div className="modal-actions">
          <button type="button" className="btn-ghost" onClick={onClose}>Cancel</button>
          <button type="submit" className="btn-accent" disabled={busy || !name.trim()}>
            {busy ? 'Creating…' : 'Create channel'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
