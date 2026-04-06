import { useState, useEffect } from 'react'
import { api } from '../lib/api'

export default function AdminSubscribers() {
  const [subscribers, setSubscribers] = useState([])

  useEffect(() => { load() }, [])

  async function load() {
    try { setSubscribers(await api.getSubscribers()) } catch (e) { console.error(e) }
  }

  async function handleDelete(id) {
    if (!confirm('Remove this subscriber?')) return
    try {
      await api.deleteSubscriber(id)
      load()
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Subscribers</h1>

      <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
        {subscribers.length > 0 ? (
          <>
            <div className="px-6 py-4 border-b border-gray-100">
              <p className="text-sm text-gray-500">{subscribers.length} subscriber{subscribers.length !== 1 ? 's' : ''}</p>
            </div>
            <div className="divide-y divide-gray-100">
              {subscribers.map(s => (
                <div key={s.id} className="px-6 py-4 flex items-center justify-between">
                  <div>
                    <span className="font-medium text-gray-900">{s.email}</span>
                    <p className="text-xs text-gray-400 mt-0.5">
                      Subscribed {new Date(s.created_at).toLocaleDateString()}
                      {s.verified && <span className="ml-2 text-emerald-600">Verified</span>}
                    </p>
                  </div>
                  <button onClick={() => handleDelete(s.id)} className="text-sm text-red-500 hover:text-red-700">
                    Remove
                  </button>
                </div>
              ))}
            </div>
          </>
        ) : (
          <div className="px-6 py-12 text-center text-gray-400">
            No subscribers yet. Users can subscribe from the public status page.
          </div>
        )}
      </div>
    </div>
  )
}
