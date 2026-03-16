import { useState, useEffect, useCallback } from 'react'
import { authApi, ApiError } from '../api/client'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  token: string | null
  loading: boolean
}

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: localStorage.getItem('token'),
    loading: true,
  })

  const fetchMe = useCallback(async () => {
    const token = localStorage.getItem('token')
    if (!token) {
      setState({ user: null, token: null, loading: false })
      return
    }
    try {
      const user = await authApi.me()
      setState({ user, token, loading: false })
    } catch {
      localStorage.removeItem('token')
      setState({ user: null, token: null, loading: false })
    }
  }, [])

  useEffect(() => {
    fetchMe()
  }, [fetchMe])

  const login = async (email: string, password: string) => {
    const { token, user } = await authApi.login(email, password)
    localStorage.setItem('token', token)
    setState({ user, token, loading: false })
  }

  const register = async (email: string, password: string) => {
    const { token, user } = await authApi.register(email, password)
    localStorage.setItem('token', token)
    setState({ user, token, loading: false })
  }

  const logout = () => {
    localStorage.removeItem('token')
    setState({ user: null, token: null, loading: false })
  }

  return { ...state, login, register, logout }
}

// Re-export ApiError for convenience
export { ApiError }
