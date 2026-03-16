export interface User {
  id: string
  email: string
  plan: 'free' | 'pro' | 'agency'
  created_at: string
}

export interface Monitor {
  id: string
  user_id: string
  name: string
  url: string
  interval_seconds: number
  status: 'up' | 'down' | 'pending'
  alert_email: boolean
  is_public: boolean
  consecutive_failures: number
  created_at: string
  updated_at: string
}

export interface Check {
  id: string
  monitor_id: string
  status_code: number | null
  response_time_ms: number | null
  is_up: boolean
  error?: string
  checked_at: string
}

export interface MonitorStats {
  uptime_1d: number
  uptime_7d: number
  uptime_30d: number
}

export interface Incident {
  id: string
  monitor_id: string
  started_at: string
  resolved_at?: string
  error?: string
}
