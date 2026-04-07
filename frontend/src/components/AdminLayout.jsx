import { useState, useEffect } from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

const navItems = [
  { to: '/admin', label: 'Dashboard', end: true },
  { to: '/admin/components', label: 'Components' },
  { to: '/admin/incidents', label: 'Incidents' },
  { to: '/admin/subscribers', label: 'Subscribers' },
]

export default function AdminLayout() {
  const navigate = useNavigate()
  const [license, setLicense] = useState(null)
  const [updates, setUpdates] = useState(null)

  useEffect(() => {
    api.getLicense().then(setLicense).catch(() => {})
    api.getUpdates().then(setUpdates).catch(() => {})
    // Poll for updates every 5 minutes
    const interval = setInterval(() => {
      api.getLicense().then(setLicense).catch(() => {})
      api.getUpdates().then(setUpdates).catch(() => {})
    }, 5 * 60 * 1000)
    return () => clearInterval(interval)
  }, [])

  function logout() {
    localStorage.removeItem('token')
    navigate('/admin/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* License expired warning */}
      {license?.expired && (
        <div className="bg-red-600 text-white px-4 py-3 text-center text-sm font-medium">
          Your license has expired. Some features may be unavailable. Please contact your vendor to renew.
        </div>
      )}

      {/* Update available banner */}
      {updates?.available && (
        <div className="bg-indigo-600 text-white px-4 py-3 text-center text-sm font-medium">
          A new version is available: {updates.updates[0]?.versionLabel}.
        </div>
      )}

      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center gap-8">
              <NavLink to="/admin" className="text-xl font-bold text-gray-900 tracking-tight">
                StatusPage
              </NavLink>
              <div className="flex gap-1">
                {navItems.map(({ to, label, end }) => (
                  <NavLink
                    key={to}
                    to={to}
                    end={end}
                    className={({ isActive }) =>
                      `px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                        isActive
                          ? 'bg-indigo-50 text-indigo-700'
                          : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                      }`
                    }
                  >
                    {label}
                  </NavLink>
                ))}
              </div>
            </div>
            <div className="flex items-center gap-4">
              {license?.available && (
                <span className="text-xs text-gray-400">
                  {license.customerName} ({license.licenseType})
                </span>
              )}
              <a
                href="/"
                target="_blank"
                className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
              >
                View Status Page
              </a>
              <button
                onClick={logout}
                className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet context={{ license }} />
      </main>
    </div>
  )
}
