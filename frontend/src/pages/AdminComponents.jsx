import { useState, useEffect } from 'react'
import { api } from '../lib/api'
import { StatusBadge } from '../components/StatusBadge'

const statuses = ['operational', 'degraded', 'major_outage', 'maintenance']

export default function AdminComponents() {
  const [components, setComponents] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState({ name: '', description: '', status: 'operational', group_name: '' })

  useEffect(() => { load() }, [])

  async function load() {
    try { setComponents(await api.getComponents()) } catch (e) { console.error(e) }
  }

  function resetForm() {
    setForm({ name: '', description: '', status: 'operational', group_name: '' })
    setEditing(null)
    setShowForm(false)
  }

  function startEdit(c) {
    setForm({ name: c.name, description: c.description, status: c.status, group_name: c.group_name })
    setEditing(c.id)
    setShowForm(true)
  }

  async function handleSubmit(e) {
    e.preventDefault()
    try {
      if (editing) {
        await api.updateComponent(editing, form)
      } else {
        await api.createComponent(form)
      }
      resetForm()
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  async function handleDelete(id) {
    if (!confirm('Delete this component?')) return
    try {
      await api.deleteComponent(id)
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  async function quickStatus(id, component, status) {
    try {
      await api.updateComponent(id, { ...component, status })
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Components</h1>
        <button
          onClick={() => { resetForm(); setShowForm(true) }}
          className="px-4 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm"
        >
          Add Component
        </button>
      </div>

      {/* Form */}
      {showForm && (
        <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {editing ? 'Edit Component' : 'New Component'}
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Name</label>
              <input
                value={form.name}
                onChange={e => setForm({ ...form, name: e.target.value })}
                required
                className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Group</label>
              <input
                value={form.group_name}
                onChange={e => setForm({ ...form, group_name: e.target.value })}
                placeholder="e.g. Infrastructure"
                className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Description</label>
            <input
              value={form.description}
              onChange={e => setForm({ ...form, description: e.target.value })}
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Status</label>
            <select
              value={form.status}
              onChange={e => setForm({ ...form, status: e.target.value })}
              className="w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
            >
              {statuses.map(s => (
                <option key={s} value={s}>{s.replace('_', ' ')}</option>
              ))}
            </select>
          </div>
          <div className="flex gap-3">
            <button type="submit" className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm">
              {editing ? 'Save Changes' : 'Create'}
            </button>
            <button type="button" onClick={resetForm} className="px-5 py-2.5 bg-white border border-gray-200 text-sm font-medium text-gray-700 rounded-xl hover:bg-gray-50 transition-colors">
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* List */}
      <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
        {components.length > 0 ? (
          <div className="divide-y divide-gray-100">
            {components.map(c => (
              <div key={c.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-medium text-gray-900">{c.name}</span>
                    {c.group_name && (
                      <span className="text-xs text-gray-400 ml-2">{c.group_name}</span>
                    )}
                    {c.description && (
                      <p className="text-sm text-gray-500 mt-0.5">{c.description}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-3">
                    <StatusBadge status={c.status} />
                    <div className="flex gap-1">
                      {statuses.map(s => (
                        <button
                          key={s}
                          onClick={() => quickStatus(c.id, c, s)}
                          title={s.replace('_', ' ')}
                          className={`w-3 h-3 rounded-full transition-transform hover:scale-125 ${
                            s === 'operational' ? 'bg-emerald-500' :
                            s === 'degraded' ? 'bg-yellow-500' :
                            s === 'major_outage' ? 'bg-red-500' : 'bg-blue-500'
                          } ${c.status === s ? 'ring-2 ring-offset-2 ring-gray-300' : 'opacity-40 hover:opacity-70'}`}
                        />
                      ))}
                    </div>
                    <button onClick={() => startEdit(c)} className="text-sm text-indigo-600 hover:text-indigo-700">Edit</button>
                    <button onClick={() => handleDelete(c.id)} className="text-sm text-red-500 hover:text-red-700">Delete</button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="px-6 py-12 text-center text-gray-400">
            No components yet. Add your first one above.
          </div>
        )}
      </div>
    </div>
  )
}
