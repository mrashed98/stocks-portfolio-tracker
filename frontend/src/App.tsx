import { Routes, Route } from 'react-router-dom'
import { Toaster } from '@/components/ui/toaster'
import Layout from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import Strategies from '@/pages/Strategies'
import ChartPage from '@/pages/ChartPage'

function App() {
  return (
    <div className="min-h-screen bg-background">
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="strategies" element={<Strategies />} />
          <Route path="portfolios" element={<div>Portfolios Page</div>} />
          <Route path="stocks" element={<div>Stocks Page</div>} />
          <Route path="charts" element={<ChartPage />} />
        </Route>
      </Routes>
      <Toaster />
    </div>
  )
}

export default App