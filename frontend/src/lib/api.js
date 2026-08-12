const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:7777'
const TOKEN_KEY = 'relay_token'
const USER_KEY = 'relay_user'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function getStoredUser() {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? JSON.parse(raw) : null
}

export function setSession(token, user) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function wsUrl(channelID) {
  const base = API_URL.replace(/^http/, 'ws')
  return `${base}/ws/${channelID}?token=${encodeURIComponent(getToken() || '')}`
}

class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.status = status
  }
}

async function request(path, { method = 'GET', body, auth = true, query } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (auth) {
    const token = getToken()
    if (token) headers.Authorization = `Bearer ${token}`
  }

  let url = `${API_URL}${path}`
  if (query) {
    const qs = new URLSearchParams(
      Object.entries(query).filter(([, v]) => v !== undefined && v !== null)
    ).toString()
    if (qs) url += `?${qs}`
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  let json
  try {
    json = await res.json()
  } catch {
    throw new ApiError('Server returned an unreadable response', res.status)
  }

  if (!res.ok || json.success === false) {
    throw new ApiError(json.error || 'Request failed', res.status)
  }

  return json.data ?? json
}

// ---- Auth ----
export const registerUser = (username, email, password) =>
  request('/api/v1/auth/register', { method: 'POST', auth: false, body: { username, email, password } })

export const loginUser = (email, password) =>
  request('/api/v1/auth/login', { method: 'POST', auth: false, body: { email, password } })

// ---- Servers ----
export const getMyServers = () => request('/api/v1/servers')

export const createServer = (name, description) =>
  request('/api/v1/servers', { method: 'POST', body: { name, description } })

export const joinServer = (inviteCode) =>
  request('/api/v1/servers/join', { method: 'POST', body: { invite_code: inviteCode } })

export const getPresence = (serverID) => request(`/api/v1/servers/${serverID}/presence`)

// ---- Channels ----
export const getChannels = (serverID) => request(`/api/v1/servers/${serverID}/channels`)

export const createChannel = (serverID, name, topic, type = 'text', isPrivate = false) =>
  request(`/api/v1/servers/${serverID}/channels`, {
    method: 'POST',
    body: { name, topic, type, is_private: isPrivate },
  })

// ---- Messages ----
export const getMessages = (channelID, limit = 50, offset = 0) =>
  request(`/api/v1/channels/${channelID}/messages`, { query: { limit, offset } })

export const sendMessage = (channelID, content) =>
  request(`/api/v1/channels/${channelID}/messages`, {
    method: 'POST',
    body: { content, type: 'text' },
  })

export { ApiError }
