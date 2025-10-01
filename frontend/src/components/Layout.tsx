import { Outlet } from 'react-router-dom'
import { Navigation } from '@/components/Navigation'

export default function Layout() {
  return (
    <div className="flex h-screen">
      <Navigation />
      <main className="flex-1 overflow-auto">
        <div className="container mx-auto p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}