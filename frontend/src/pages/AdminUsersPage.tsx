import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { adminApi, ApiError } from '../api/client'
import { useAuth } from '../hooks/useAuth'
import ErrorMessage from '../components/ErrorMessage'
import type { User } from '../api/types'

export default function AdminUsersPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [users, setUsers] = useState<User[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchUsers() {
      try {
        const data = await adminApi.listUsers()
        setUsers(data ?? [])
      } catch (err) {
        setError(err instanceof ApiError ? err.message : 'Failed to load users')
      } finally {
        setLoading(false)
      }
    }
    fetchUsers()
  }, [])

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800 bg-gray-900">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-bold text-white">u-status</h1>
            <span className="text-xs px-2 py-0.5 rounded bg-red-900 text-red-200 uppercase tracking-wide">Admin</span>
          </div>
          <div className="flex items-center gap-4">
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
          <h2 className="text-xl font-semibold text-white">All Users</h2>
          <p className="text-sm text-gray-400 mt-0.5">{users.length} user{users.length !== 1 ? 's' : ''}</p>
        </div>

        {error && <ErrorMessage message={error} />}

        {loading ? (
          <div className="flex justify-center py-20">
            <div className="w-6 h-6 border-2 border-green-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : users.length === 0 && !error ? (
          <div className="text-center py-20 text-gray-500">
            <p className="text-lg">No users found</p>
          </div>
        ) : (
          <div className="space-y-2">
            {users.map((u) => (
              <Link
                key={u.id}
                to={`/admin/users/${u.id}/monitors`}
                className="block rounded-lg border border-gray-800 bg-gray-900 px-5 py-4 hover:border-gray-600 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-white">{u.email}</p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      Joined {new Date(u.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs px-2 py-0.5 rounded bg-gray-800 text-gray-300 uppercase tracking-wide">
                      {u.plan}
                    </span>
                    {u.is_admin && (
                      <span className="text-xs px-2 py-0.5 rounded bg-red-900 text-red-200 uppercase tracking-wide">
                        Admin
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
