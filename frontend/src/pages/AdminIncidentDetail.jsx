import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import { StatusBadge, ImpactBadge } from '../components/StatusBadge'

const statuses = ['investigating', 'identified', 'monitoring', 'resolved']
const impacts = ['none', 'minor', 'major', 'critical']

export default function AdminIncidentDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [incident, setIncident] = useState(null)
  const [loading, setLoading] = useState(true)
  const [updateForm, setUpdateForm] = useState({ status: '', message: '' })
  const [statusForm, setStatusForm] = useState({ status: '', impact: '' })

  useEffect(() => { load() }, [id])

  async function load() {
    try {
      const data = await api.getIncident(id)
      setIncident(data)
      setStatusForm({ status: data.status, impact: data.impact })
      setUpdateForm(prev => ({ ...prev, status: data.status }))
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  async function handleStatusUpdate(e) {
    e.preventDefault()
    try {
      await api.updateIncident(id, statusForm)
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  async function handlePostUpdate(e) {
    e.preventDefault()
    try {
      await api.createIncidentUpdate(id, updateForm)
      setUpdateForm({ status: incident.status, message: '' })
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  async function handleDelete() {
    if (!confirm('Delete this incident?')) return
    try {
      await api.deleteIncident(id)
      navigate('/admin/incidents')
    } catch (err) {
      alert(err.message)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
      </div>
    )
  }

  if (!incident) {
    return <div className="text-center py-20 text-gray-400">Incident not found</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{incident.title}</h1>
          <div className="flex gap-2 mt-2">
            <StatusBadge status={incident.status} />
            <ImpactBadge impact={incident.impact} />
            <span className="text-sm text-gray-400">Created {new Date(incident.created_at).toLocaleString()}</span>
          </div>
        </div>
        <button onClick={handleDelete} className="text-sm text-red-500 hover:text-red-700">
          Delete Incident
        </button>
      </div>

      {/* Update Status */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <form onSubmit={handleStatusUpdate} className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Update Status</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Status</label>
            <select
              value={statusForm.status}
              onChange={e => setStatusForm({ ...statusForm, status: e.target.value })}
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
            >
              {statuses.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Impact</label>
            <select
              value={statusForm.impact}
              onChange={e => setStatusForm({ ...statusForm, impact: e.target.value })}
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
            >
              {impacts.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <button type="submit" className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm">
            Update
          </button>
        </form>

        {/* Post Update */}
        <form onSubmit={handlePostUpdate} className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Post Update</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Status</label>
            <select
              value={updateForm.status}
              onChange={e => setUpdateForm({ ...updateForm, status: e.target.value })}
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
            >
              {statuses.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Message</label>
            <textarea
              value={updateForm.message}
              onChange={e => setUpdateForm({ ...updateForm, message: e.target.value })}
              required
              rows={3}
              placeholder="Describe the current situation..."
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
            />
          </div>
          <button type="submit" className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm">
            Post Update
          </button>
        </form>
      </div>

      {/* Timeline */}
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
                    <div className="absolute left-0 top-1.5 w-4 h-4 rounded-full bg-white border-2 border-indigo-400" />
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
  )
}
