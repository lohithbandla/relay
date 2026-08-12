import { createContext, useContext, useState, useCallback } from 'react'
import { loginUser, registerUser, setSession, clearSession, getStoredUser, getToken } from '../lib/api'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(getStoredUser)
  const [token, setToken] = useState(getToken)

  const login = useCallback(async (email, password) => {
    const data = await loginUser(email, password)
    setSession(data.token, data.user)
    setUser(data.user)
    setToken(data.token)
    return data.user
  }, [])

  const register = useCallback(async (username, email, password) => {
    const data = await registerUser(username, email, password)
    setSession(data.token, data.user)
    setUser(data.user)
    setToken(data.token)
    return data.user
  }, [])

  const logout = useCallback(() => {
    clearSession()
    setUser(null)
    setToken(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, token, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
