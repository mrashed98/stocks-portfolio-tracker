import { Routes, Route } from 'react-router-dom'
import { Toaster } from '@/components/ui/toaster'
import { AuthProvider } from '@/contexts/AuthContext'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import Layout from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import Strategies from '@/pages/Strategies'
import Portfolios from '@/pages/Portfolios'
import Stocks from '@/pages/Stocks'
import ChartPage from '@/pages/ChartPage'
import Login from '@/pages/Login'
import Register from '@/pages/Register'

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <div className="min-h-screen bg-background">
          <Routes>
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            
            {/* Protected routes */}
            <Route path="/" element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }>
              <Route index element={<Dashboard />} />
              <Route path="strategies" element={<Strategies />} />
              <Route path="portfolios" element={<Portfolios />} />
              <Route path="stocks" element={<Stocks />} />
              <Route path="charts" element={<ChartPage />} />
            </Route>
          </Routes>
          <Toaster />
        </div>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App