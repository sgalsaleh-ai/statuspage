import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { connectCentrifugo } from '../lib/centrifuge'
import { StatusBadge, ImpactBadge } from '../components/StatusBadge'

const overallLabels = {
  operational: { text: 'All Systems Operational', bg: 'bg-emerald-500', icon: '✓' },
  degraded: { text: 'Degraded Performance', bg: 'bg-yellow-500', icon: '!' },
  major_outage: { text: 'Major System Outage', bg: 'bg-red-500', icon: '✕' },
  maintenance: { text: 'Scheduled Maintenance', bg: 'bg-blue-500', icon: '⚙' },
}

export default function PublicStatus() {
  const [status, setStatus] = useState(null)
  const [loading, setLoading] = useState(true)
  const [email, setEmail] = useState('')
  const [subscribeMsg, setSubscribeMsg] = useState('')

  useEffect(() => {
    loadStatus()
    let disconnect
    api.getCentrifugoConfig().then(config => {
      disconnect = connectCentrifugo(config.url, () => loadStatus())
    }).catch(() => {})
    return () => disconnect?.()
  }, [])

  async function loadStatus() {
    try {
      const data = await api.getStatus()
      setStatus(data)
    } catch (e) {
      console.error('Failed to load status:', e)
    } finally {
      setLoading(false)
    }
  }

  async function handleSubscribe(e) {
    e.preventDefault()
    try {
      await api.subscribe(email)
      setSubscribeMsg('Subscribed successfully!')
      setEmail('')
    } catch (err) {
      setSubscribeMsg(err.message)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
      </div>
    )
  }

  const overall = overallLabels[status?.overall_status] || overallLabels.operational

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-3xl mx-auto px-4 py-6">
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">StatusPage</h1>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-8 space-y-6">
        {/* Overall Status Banner */}
        <div className={`${overall.bg} rounded-2xl p-6 text-white shadow-sm`}>
          <div className="flex items-center gap-3">
            <span className="text-2xl">{overall.icon}</span>
            <span className="text-lg font-semibold">{overall.text}</span>
          </div>
        </div>

        {/* Active Incidents */}
        {status?.active_incidents?.length > 0 && (
          <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-100">
              <h2 className="text-lg font-semibold text-gray-900">Active Incidents</h2>
            </div>
            <div className="divide-y divide-gray-100">
              {status.active_incidents.map(incident => (
                <Link
                  key={incident.id}
                  to={`/incident/${incident.id}`}
                  className="block px-6 py-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <h3 className="font-medium text-gray-900">{incident.title}</h3>
                      <p className="text-sm text-gray-500 mt-1">
                        {new Date(incident.created_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <StatusBadge status={incident.status} />
                      <ImpactBadge impact={incident.impact} />
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}

        {/* Components */}
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900">Components</h2>
          </div>
          {status?.components?.length > 0 ? (
            <div className="divide-y divide-gray-100">
              {status.components.map(component => (
                <div key={component.id} className="px-6 py-4 flex items-center justify-between">
                  <div>
                    <span className="font-medium text-gray-900">{component.name}</span>
                    {component.description && (
                      <p className="text-sm text-gray-500 mt-0.5">{component.description}</p>
                    )}
                  </div>
                  <StatusBadge status={component.status} />
                </div>
              ))}
            </div>
          ) : (
            <div className="px-6 py-12 text-center text-gray-400">
              No components configured yet.
            </div>
          )}
        </div>

        {/* Subscribe */}
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Subscribe to Updates</h2>
          <p className="text-sm text-gray-500 mb-4">Get notified when incidents are created or resolved.</p>
          <form onSubmit={handleSubscribe} className="flex gap-3">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              required
              className="flex-1 rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
            <button
              type="submit"
              className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm"
            >
              Subscribe
            </button>
          </form>
          {subscribeMsg && (
            <p className="text-sm mt-3 text-gray-600">{subscribeMsg}</p>
          )}
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-gray-400 pt-4">
          Powered by StatusPage
        </p>
      </div>
    </div>
  )
}
