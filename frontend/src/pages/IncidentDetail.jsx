import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../lib/api'
import { StatusBadge, ImpactBadge } from '../components/StatusBadge'

export default function IncidentDetail() {
  const { id } = useParams()
  const [incident, setIncident] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getIncident(id).then(setIncident).catch(console.error).finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-600" />
      </div>
    )
  }

  if (!incident) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Incident not found</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-3xl mx-auto px-4 py-6">
          <Link to="/" className="text-sm text-emerald-600 hover:text-emerald-700 mb-2 inline-block">&larr; Back to status</Link>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">{incident.title}</h1>
          <div className="flex gap-2 mt-3">
            <StatusBadge status={incident.status} />
            <ImpactBadge impact={incident.impact} />
          </div>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-8">
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900">Timeline</h2>
          </div>
          {incident.updates?.length > 0 ? (
            <div className="px-6 py-4">
              <div className="relative">
                <div className="absolute left-[7px] top-2 bottom-2 w-0.5 bg-gray-200" />
                <div className="space-y-6">
                  {incident.updates.map(update => (
                    <div key={update.id} className="relative pl-8">
                      <div className="absolute left-0 top-1.5 w-4 h-4 rounded-full bg-white border-2 border-emerald-400" />
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <StatusBadge status={update.status} />
                          <span className="text-xs text-gray-400">
                            {new Date(update.created_at).toLocaleString()}
                          </span>
                        </div>
                        <p className="text-sm text-gray-700">{update.message}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <div className="px-6 py-12 text-center text-gray-400">No updates yet.</div>
          )}
        </div>
      </div>
    </div>
  )
}
