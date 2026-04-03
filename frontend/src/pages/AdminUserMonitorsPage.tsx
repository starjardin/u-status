import { useState, useEffect } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { adminApi, ApiError } from '../api/client'
import { useAuth } from '../hooks/useAuth'
import StatusBadge from '../components/StatusBadge'
import ErrorMessage from '../components/ErrorMessage'
import type { Monitor } from '../api/types'

export default function AdminUserMonitorsPage() {
  const { userId } = useParams<{ userId: string }>()
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [monitors, setMonitors] = useState<Monitor[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!userId) return
    async function fetchMonitors() {
      try {
        const data = await adminApi.getUserMonitors(userId!)
        setMonitors(data ?? [])
      } catch (err) {
        setError(err instanceof ApiError ? err.message : 'Failed to load monitors')
      } finally {
        setLoading(false)
      }
    }
    fetchMonitors()
  }, [userId])

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800 bg-gray-900">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-bold text-white">u-status</h1>
            <span className="text-xs px-2 py-0.5 rounded bg-red-900 text-red-200 uppercase tracking-wide">Admin</span>
          </div>
          <div className="flex items-center gap-4">
            <Link to="/admin/users" className="text-sm text-gray-400 hover:text-white transition-colors">
              All Users
            </Link>
            <Link to="/dashboard" className="text-sm text-gray-400 hover:text-white transition-colors">
              Dashboard
            </Link>
            <span className="text-sm text-gray-400">{user?.email}</span>
            <button
              onClick={() => { logout(); navigate('/login') }}
              className="text-sm text-gray-400 hover:text-white transition-colors"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-8">
        <div className="mb-6">
          <Link to="/admin/users" className="text-sm text-gray-400 hover:text-white transition-colors">
            &larr; Back to users
          </Link>
          <h2 className="text-xl font-semibold text-white mt-2">User Monitors</h2>
          <p className="text-sm text-gray-400 mt-0.5">{monitors.length} monitor{monitors.length !== 1 ? 's' : ''}</p>
        </div>

        {error && <ErrorMessage message={error} />}

        {loading ? (
          <div className="flex justify-center py-20">
            <div className="w-6 h-6 border-2 border-green-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : monitors.length === 0 && !error ? (
          <div className="text-center py-20 text-gray-500">
            <p className="text-lg">This user has no monitors</p>
          </div>
        ) : (
          <div className="space-y-2">
            {monitors.map((m) => (
              <div
                key={m.id}
                className="rounded-lg border border-gray-800 bg-gray-900 px-5 py-4 flex items-center gap-4"
              >
                <StatusBadge status={m.status} />
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-white">{m.name}</p>
                  <p className="text-xs text-gray-500 truncate mt-0.5">{m.url}</p>
                </div>
                <div className="text-xs text-gray-400 hidden sm:block">
                  Every {m.interval_seconds}s
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
