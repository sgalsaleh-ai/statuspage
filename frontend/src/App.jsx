import { Routes, Route, Navigate } from 'react-router-dom'
import PublicStatus from './pages/PublicStatus'
import IncidentDetail from './pages/IncidentDetail'
import Login from './pages/Login'
import Setup from './pages/Setup'
import AdminLayout from './components/AdminLayout'
import AdminDashboard from './pages/AdminDashboard'
import AdminComponents from './pages/AdminComponents'
import AdminIncidents from './pages/AdminIncidents'
import AdminIncidentDetail from './pages/AdminIncidentDetail'
import AdminSubscribers from './pages/AdminSubscribers'

function ProtectedRoute({ children }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/admin/login" replace />
  return children
}

export default function App() {
  return (
    <Routes>
      {/* Public */}
      <Route path="/" element={<PublicStatus />} />
      <Route path="/incident/:id" element={<IncidentDetail />} />

      {/* Auth */}
      <Route path="/admin/login" element={<Login />} />
      <Route path="/admin/setup" element={<Setup />} />

      {/* Admin */}
      <Route path="/admin" element={<ProtectedRoute><AdminLayout /></ProtectedRoute>}>
        <Route index element={<AdminDashboard />} />
        <Route path="components" element={<AdminComponents />} />
        <Route path="incidents" element={<AdminIncidents />} />
        <Route path="incidents/:id" element={<AdminIncidentDetail />} />
        <Route path="subscribers" element={<AdminSubscribers />} />
      </Route>
    </Routes>
  )
}
