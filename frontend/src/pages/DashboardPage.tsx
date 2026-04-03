import { useState, useCallback, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { monitorsApi, ApiError } from '../api/client'
import { useAuth } from '../hooks/useAuth'
import { useInterval } from '../hooks/useInterval'
import StatusBadge from '../components/StatusBadge'
import ErrorMessage from '../components/ErrorMessage'
import type { Monitor } from '../api/types'

export default function DashboardPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [monitors, setMonitors] = useState<Monitor[]>([])
  const [loadError, setLoadError] = useState('')
  const [showAddForm, setShowAddForm] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [formName, setFormName] = useState('')
  const [formUrl, setFormUrl] = useState('')
  const [formInterval, setFormInterval] = useState('60')
  const [formError, setFormError] = useState('')
  const [formLoading, setFormLoading] = useState(false)

  const fetchMonitors = useCallback(async () => {
    try {
      const data = await monitorsApi.list()
      setMonitors(data ?? [])
      setLoadError('')
    } catch (err) {
      setLoadError(err instanceof ApiError ? err.message : 'Failed to load monitors')
    }
  }, [])

  // Initial load + poll every 30s
  useState(() => { fetchMonitors() })
  useInterval(fetchMonitors, 30_000)

  const handleAdd = async (e: FormEvent) => {
    e.preventDefault()
    const name = formName.trim()
    const url = formUrl.trim()
    if (!name) { setFormError('Name is required'); return }
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      setFormError('URL must start with http:// or https://')
      return
    }
    const interval = parseInt(formInterval, 10)
    if (isNaN(interval) || interval < 30) {
      setFormError('Interval must be at least 30 seconds')
      return
    }

    setFormError('')
    setFormLoading(true)
    try {
      await monitorsApi.create(name, url, interval)
      setFormName('')
      setFormUrl('')
      setFormInterval('60')
      setShowAddForm(false)
      fetchMonitors()
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : 'Failed to add monitor')
    } finally {
      setFormLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await monitorsApi.delete(id)
      setDeleteConfirm(null)
      fetchMonitors()
    } catch (err) {
      setLoadError(err instanceof ApiError ? err.message : 'Failed to delete monitor')
    }
  }

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <h1 className="text-lg font-bold text-white">u-status</h1>
          <div className="flex items-center gap-4">
            {user?.is_admin && (
              <Link to="/admin/users" className="text-sm text-red-400 hover:text-red-300 font-medium transition-colors">
                Admin
              </Link>
            )}
            <span className="text-sm text-gray-400">{user?.email}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-gray-800 text-gray-300 uppercase tracking-wide">
              {user?.plan}
            </span>
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
        {/* Title row */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-xl font-semibold text-white">Monitors</h2>
            <p className="text-sm text-gray-400 mt-0.5">{monitors.length} monitor{monitors.length !== 1 ? 's' : ''}</p>
          </div>
          <button
            onClick={() => setShowAddForm(true)}
            className="px-4 py-2 rounded-md bg-green-600 hover:bg-green-500 text-white text-sm font-semibold transition-colors"
          >
            + Add Monitor
          </button>
        </div>

        {loadError && <ErrorMessage message={loadError} />}

        {/* Add form */}
        {showAddForm && (
          <div className="mb-6 rounded-lg border border-gray-700 bg-gray-900 p-5">
            <h3 className="text-sm font-semibold text-white mb-4">Add Monitor</h3>
            <form onSubmit={handleAdd} className="space-y-3">
              {formError && <ErrorMessage message={formError} />}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <input
                  type="text"
                  placeholder="Name"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="rounded-md bg-gray-800 border border-gray-700 px-3 py-2 text-white placeholder-gray-500 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                />
                <input
                  type="url"
                  placeholder="https://example.com"
                  value={formUrl}
                  onChange={(e) => setFormUrl(e.target.value)}
                  className="rounded-md bg-gray-800 border border-gray-700 px-3 py-2 text-white placeholder-gray-500 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                />
                <select
                  value={formInterval}
                  onChange={(e) => setFormInterval(e.target.value)}
                  className="rounded-md bg-gray-800 border border-gray-700 px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                >
                  <option value="30">Every 30s</option>
                  <option value="60">Every 60s</option>
                  <option value="300">Every 5m</option>
                  <option value="600">Every 10m</option>
                </select>
              </div>
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={formLoading}
                  className="px-4 py-1.5 rounded-md bg-green-600 hover:bg-green-500 disabled:opacity-50 text-white text-sm font-semibold transition-colors"
                >
                  {formLoading ? 'Adding…' : 'Add'}
                </button>
                <button
                  type="button"
                  onClick={() => { setShowAddForm(false); setFormError('') }}
                  className="px-4 py-1.5 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Monitor list */}
        {monitors.length === 0 && !loadError ? (
          <div className="text-center py-20 text-gray-500">
            <p className="text-lg">No monitors yet</p>
            <p className="text-sm mt-1">Add a URL above to start monitoring</p>
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
                  <Link
                    to={`/dashboard/monitors/${m.id}`}
                    className="font-medium text-white hover:text-green-400 transition-colors"
                  >
                    {m.name}
                  </Link>
                  <p className="text-xs text-gray-500 truncate mt-0.5">{m.url}</p>
                </div>
                <div className="text-xs text-gray-400 hidden sm:block">
                  Every {m.interval_seconds}s
                </div>
                {deleteConfirm === m.id ? (
                  <div className="flex items-center gap-2 text-sm">
                    <span className="text-gray-400">Delete?</span>
                    <button
                      onClick={() => handleDelete(m.id)}
                      className="text-red-400 hover:text-red-300 font-medium"
                    >
                      Yes
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(null)}
                      className="text-gray-400 hover:text-gray-300"
                    >
                      No
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setDeleteConfirm(m.id)}
                    className="text-gray-600 hover:text-red-400 transition-colors text-sm"
                    title="Delete monitor"
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
