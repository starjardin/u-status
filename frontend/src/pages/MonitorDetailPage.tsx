import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts'
import { monitorsApi, ApiError } from '../api/client'
import StatusBadge from '../components/StatusBadge'
import ErrorMessage from '../components/ErrorMessage'
import type { Monitor, Check, MonitorStats, Incident } from '../api/types'

function formatDuration(start: string, end?: string): string {
  const ms = new Date(end ?? new Date()).getTime() - new Date(start).getTime()
  const secs = Math.floor(ms / 1000)
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.floor(secs / 60)}m ${secs % 60}s`
  return `${Math.floor(secs / 3600)}h ${Math.floor((secs % 3600) / 60)}m`
}

function formatTime(dt: string): string {
  return new Date(dt).toLocaleString(undefined, {
    month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function UptimePct({ label, value }: { label: string; value: number }) {
  const color = value >= 99 ? 'text-green-400' : value >= 95 ? 'text-yellow-400' : 'text-red-400'
  return (
    <div className="rounded-md bg-gray-900 border border-gray-800 px-4 py-3 text-center">
      <div className={`text-2xl font-bold ${color}`}>{value.toFixed(2)}%</div>
      <div className="text-xs text-gray-500 mt-1">{label}</div>
    </div>
  )
}

export default function MonitorDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [monitor, setMonitor] = useState<Monitor | null>(null)
  const [checks, setChecks] = useState<Check[]>([])
  const [stats, setStats] = useState<MonitorStats | null>(null)
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    Promise.all([
      monitorsApi.list(),
      monitorsApi.getChecks(id, 24),
      monitorsApi.getStats(id),
      monitorsApi.getIncidents(id),
    ])
      .then(([monitorsList, checksData, statsData, incidentsData]) => {
        const m = monitorsList?.find((x) => x.id === id)
        if (!m) { setError('Monitor not found'); return }
        setMonitor(m)
        setChecks(checksData ?? [])
        setStats(statsData)
        setIncidents(incidentsData ?? [])
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : 'Failed to load monitor')
      })
      .finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-950">
        <div className="w-6 h-6 border-2 border-green-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  if (error || !monitor) {
    return (
      <div className="min-h-screen bg-gray-950 p-8">
        <ErrorMessage message={error || 'Monitor not found'} />
        <Link to="/dashboard" className="mt-4 inline-block text-green-400 hover:text-green-300 text-sm">
          ← Back to dashboard
        </Link>
      </div>
    )
  }

  // Prepare chart data — reversed so oldest is left
  const chartData = [...checks].reverse().map((c) => ({
    time: new Date(c.checked_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    ms: c.response_time_ms ?? 0,
    up: c.is_up,
  }))

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 py-4">
          <Link to="/dashboard" className="text-sm text-gray-400 hover:text-white transition-colors">
            ← Dashboard
          </Link>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-8 space-y-8">
        {/* Monitor info */}
        <div className="flex items-start gap-4">
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-white">{monitor.name}</h1>
            <a
              href={monitor.url}
              target="_blank"
              rel="noreferrer"
              className="text-sm text-gray-400 hover:text-green-400 transition-colors"
            >
              {monitor.url}
            </a>
          </div>
          <StatusBadge status={monitor.status} />
        </div>

        {/* Uptime stats */}
        {stats && (
          <div className="grid grid-cols-3 gap-3">
            <UptimePct label="Last 24h" value={stats.uptime_1d} />
            <UptimePct label="Last 7 days" value={stats.uptime_7d} />
            <UptimePct label="Last 30 days" value={stats.uptime_30d} />
          </div>
        )}

        {/* Response time chart */}
        <section>
          <h2 className="text-sm font-semibold text-gray-300 uppercase tracking-wider mb-3">
            Response Time (last 24h)
          </h2>
          {chartData.length === 0 ? (
            <div className="rounded-lg border border-gray-800 bg-gray-900 py-10 text-center text-gray-500 text-sm">
              No checks recorded yet
            </div>
          ) : (
            <div className="rounded-lg border border-gray-800 bg-gray-900 p-4">
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
                  <XAxis dataKey="time" tick={{ fill: '#6b7280', fontSize: 11 }} interval="preserveStartEnd" />
                  <YAxis tick={{ fill: '#6b7280', fontSize: 11 }} unit="ms" width={55} />
                  <Tooltip
                    contentStyle={{ background: '#111827', border: '1px solid #374151', borderRadius: 6 }}
                    labelStyle={{ color: '#9ca3af' }}
                    itemStyle={{ color: '#22c55e' }}
                    formatter={(v: number) => [`${v}ms`, 'Response time']}
                  />
                  <Line
                    type="monotone"
                    dataKey="ms"
                    stroke="#22c55e"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </section>

        {/* Incidents */}
        <section>
          <h2 className="text-sm font-semibold text-gray-300 uppercase tracking-wider mb-3">
            Incidents
          </h2>
          {incidents.length === 0 ? (
            <div className="rounded-lg border border-gray-800 bg-gray-900 py-6 text-center text-gray-500 text-sm">
              No incidents recorded 🎉
            </div>
          ) : (
            <div className="rounded-lg border border-gray-800 overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-900 text-gray-400 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-4 py-2 text-left">Started</th>
                    <th className="px-4 py-2 text-left">Resolved</th>
                    <th className="px-4 py-2 text-left">Duration</th>
                    <th className="px-4 py-2 text-left">Error</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {incidents.map((inc) => (
                    <tr key={inc.id} className="bg-gray-950">
                      <td className="px-4 py-3 text-gray-300">{formatTime(inc.started_at)}</td>
                      <td className="px-4 py-3">
                        {inc.resolved_at ? (
                          <span className="text-green-400">{formatTime(inc.resolved_at)}</span>
                        ) : (
                          <span className="text-red-400">Ongoing</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-gray-400">
                        {formatDuration(inc.started_at, inc.resolved_at)}
                      </td>
                      <td className="px-4 py-3 text-gray-500 text-xs">{inc.error ?? '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {/* Recent checks */}
        <section>
          <h2 className="text-sm font-semibold text-gray-300 uppercase tracking-wider mb-3">
            Recent Checks
          </h2>
          {checks.length === 0 ? (
            <div className="rounded-lg border border-gray-800 bg-gray-900 py-6 text-center text-gray-500 text-sm">
              No checks yet
            </div>
          ) : (
            <div className="rounded-lg border border-gray-800 overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-900 text-gray-400 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-4 py-2 text-left">Time</th>
                    <th className="px-4 py-2 text-left">Status</th>
                    <th className="px-4 py-2 text-left">Code</th>
                    <th className="px-4 py-2 text-left">Response</th>
                    <th className="px-4 py-2 text-left">Error</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {checks.slice(0, 50).map((c) => (
                    <tr key={c.id} className="bg-gray-950">
                      <td className="px-4 py-2 text-gray-400 text-xs">{formatTime(c.checked_at)}</td>
                      <td className="px-4 py-2">
                        <span className={`text-xs font-semibold ${c.is_up ? 'text-green-400' : 'text-red-400'}`}>
                          {c.is_up ? 'UP' : 'DOWN'}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-gray-400">{c.status_code ?? '—'}</td>
                      <td className="px-4 py-2 text-gray-400">
                        {c.response_time_ms != null ? `${c.response_time_ms}ms` : '—'}
                      </td>
                      <td className="px-4 py-2 text-gray-500 text-xs">{c.error ?? '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </div>
  )
}
