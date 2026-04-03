import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import MonitorDetailPage from './pages/MonitorDetailPage'
import AdminUsersPage from './pages/AdminUsersPage'
import AdminUserMonitorsPage from './pages/AdminUserMonitorsPage'
import ProtectedRoute from './components/ProtectedRoute'
import AdminRoute from './components/AdminRoute'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/dashboard/monitors/:id"
        element={
          <ProtectedRoute>
            <MonitorDetailPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/users"
        element={
          <AdminRoute>
            <AdminUsersPage />
          </AdminRoute>
        }
      />
      <Route
        path="/admin/users/:userId/monitors"
        element={
          <AdminRoute>
            <AdminUserMonitorsPage />
          </AdminRoute>
        }
      />
    </Routes>
  )
}
