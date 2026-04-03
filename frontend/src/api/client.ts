/// <reference types="vite/client" />

import type { User, Monitor, Check, MonitorStats, Incident } from './types'

const API_BASE = import.meta.env.VITE_API_URL ?? ''

function getToken(): string | null {
  return localStorage.getItem('token')
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new ApiError(res.status, body.error ?? 'Unknown error')
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// Auth
export const authApi = {
  register: (email: string, password: string) =>
    request<{ token: string; user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  me: () => request<User>('/api/me'),
}

// Monitors
export const monitorsApi = {
  list: () => request<Monitor[]>('/api/monitors/'),

  create: (name: string, url: string, intervalSeconds = 60) =>
    request<Monitor>('/api/monitors/', {
      method: 'POST',
      body: JSON.stringify({ name, url, interval_seconds: intervalSeconds }),
    }),

  update: (id: string, updates: Partial<Pick<Monitor, 'name' | 'url' | 'interval_seconds' | 'alert_email' | 'is_public'>>) =>
    request<Monitor>(`/api/monitors/${id}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    }),

  delete: (id: string) =>
    request<void>(`/api/monitors/${id}`, { method: 'DELETE' }),

  getChecks: (id: string, hours = 24) =>
    request<Check[]>(`/api/monitors/${id}/checks?hours=${hours}`),

  getStats: (id: string) =>
    request<MonitorStats>(`/api/monitors/${id}/stats`),

  getIncidents: (id: string) =>
    request<Incident[]>(`/api/monitors/${id}/incidents`),
}

// Admin
export const adminApi = {
  listUsers: () => request<User[]>('/api/admin/users'),

  getUserMonitors: (userId: string) =>
    request<Monitor[]>(`/api/admin/users/${userId}/monitors`),
}
