import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'

export default function AdminDashboard() {
  const [components, setComponents] = useState([])
  const [incidents, setIncidents] = useState([])

  useEffect(() => {
    api.getComponents().then(setComponents).catch(console.error)
    api.getAdminIncidents().then(setIncidents).catch(console.error)
  }, [])

  const activeIncidents = incidents.filter(i => i.status !== 'resolved')
  const operational = components.filter(c => c.status === 'operational').length

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm font-medium text-gray-500">Components</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{components.length}</p>
          <p className="text-sm text-emerald-600 mt-1">{operational} operational</p>
        </div>
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm font-medium text-gray-500">Active Incidents</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{activeIncidents.length}</p>
          <p className="text-sm text-gray-500 mt-1">{incidents.length} total</p>
        </div>
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm font-medium text-gray-500">Overall Status</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">
            {activeIncidents.length === 0 ? '✓' : '!'}
          </p>
          <p className="text-sm mt-1">
            {activeIncidents.length === 0 ? (
              <span className="text-emerald-600">All systems operational</span>
            ) : (
              <span className="text-yellow-600">Issues detected</span>
            )}
          </p>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="flex gap-3">
        <Link
          to="/admin/components"
          className="px-4 py-2.5 bg-white border border-gray-200 text-sm font-medium text-gray-700 rounded-xl hover:bg-gray-50 transition-colors shadow-sm"
        >
          Manage Components
        </Link>
        <Link
          to="/admin/incidents"
          className="px-4 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm"
        >
          Create Incident
        </Link>
      </div>

      {/* Recent Active Incidents */}
      {activeIncidents.length > 0 && (
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900">Active Incidents</h2>
          </div>
          <div className="divide-y divide-gray-100">
            {activeIncidents.map(incident => (
              <Link
                key={incident.id}
                to={`/admin/incidents/${incident.id}`}
                className="block px-6 py-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium text-gray-900">{incident.title}</span>
                  <span className="text-xs text-gray-500">{incident.status}</span>
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
