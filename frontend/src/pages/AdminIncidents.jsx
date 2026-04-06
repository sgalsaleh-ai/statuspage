import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { StatusBadge, ImpactBadge } from '../components/StatusBadge'

const statuses = ['investigating', 'identified', 'monitoring', 'resolved']
const impacts = ['none', 'minor', 'major', 'critical']

export default function AdminIncidents() {
  const [incidents, setIncidents] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ title: '', status: 'investigating', impact: 'none', message: '' })

  useEffect(() => { load() }, [])

  async function load() {
    try { setIncidents(await api.getAdminIncidents()) } catch (e) { console.error(e) }
  }

  async function handleSubmit(e) {
    e.preventDefault()
    try {
      await api.createIncident(form)
      setForm({ title: '', status: 'investigating', impact: 'none', message: '' })
      setShowForm(false)
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  async function handleDelete(id) {
    if (!confirm('Delete this incident?')) return
    try {
      await api.deleteIncident(id)
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  const active = incidents.filter(i => i.status !== 'resolved')
  const resolved = incidents.filter(i => i.status === 'resolved')

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Incidents</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm"
        >
          Create Incident
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">New Incident</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Title</label>
            <input
              value={form.title}
              onChange={e => setForm({ ...form, title: e.target.value })}
              required
              placeholder="e.g. API Response Times Elevated"
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Status</label>
              <select
                value={form.status}
                onChange={e => setForm({ ...form, status: e.target.value })}
                className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
              >
                {statuses.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Impact</label>
              <select
                value={form.impact}
                onChange={e => setForm({ ...form, impact: e.target.value })}
                className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
              >
                {impacts.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Initial Message</label>
            <textarea
              value={form.message}
              onChange={e => setForm({ ...form, message: e.target.value })}
              rows={3}
              placeholder="Describe what's happening..."
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
            />
          </div>
          <div className="flex gap-3">
            <button type="submit" className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm">
              Create Incident
            </button>
            <button type="button" onClick={() => setShowForm(false)} className="px-5 py-2.5 bg-white border border-gray-200 text-sm font-medium text-gray-700 rounded-xl hover:bg-gray-50 transition-colors">
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Active Incidents */}
      {active.length > 0 && (
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900">Active ({active.length})</h2>
          </div>
          <div className="divide-y divide-gray-100">
            {active.map(incident => (
              <div key={incident.id} className="px-6 py-4 flex items-center justify-between">
                <Link to={`/admin/incidents/${incident.id}`} className="flex-1">
                  <span className="font-medium text-gray-900 hover:text-indigo-600 transition-colors">{incident.title}</span>
                  <p className="text-xs text-gray-400 mt-0.5">{new Date(incident.created_at).toLocaleString()}</p>
                </Link>
                <div className="flex items-center gap-3">
                  <StatusBadge status={incident.status} />
                  <ImpactBadge impact={incident.impact} />
                  <button onClick={() => handleDelete(incident.id)} className="text-sm text-red-500 hover:text-red-700">Delete</button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Resolved */}
      {resolved.length > 0 && (
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="text-lg font-semibold text-gray-900">Resolved ({resolved.length})</h2>
          </div>
          <div className="divide-y divide-gray-100">
            {resolved.map(incident => (
              <div key={incident.id} className="px-6 py-4 flex items-center justify-between">
                <Link to={`/admin/incidents/${incident.id}`} className="flex-1">
                  <span className="font-medium text-gray-500 hover:text-indigo-600 transition-colors">{incident.title}</span>
                  <p className="text-xs text-gray-400 mt-0.5">{new Date(incident.created_at).toLocaleString()}</p>
                </Link>
                <div className="flex items-center gap-3">
                  <StatusBadge status={incident.status} />
                  <button onClick={() => handleDelete(incident.id)} className="text-sm text-red-500 hover:text-red-700">Delete</button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {incidents.length === 0 && (
        <div className="bg-white rounded-2xl border border-gray-200 shadow-sm px-6 py-12 text-center text-gray-400">
          No incidents yet.
        </div>
      )}
    </div>
  )
}
