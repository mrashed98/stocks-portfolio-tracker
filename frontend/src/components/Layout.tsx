import { Outlet } from 'react-router-dom'
import { Navigation } from '@/components/Navigation'
import { ErrorBoundary } from '@/components/ErrorBoundary'

export default function Layout() {
  return (
    <div className="flex h-screen">
      <ErrorBoundary
        fallback={
          <div className="w-64 bg-card border-r border-border flex items-center justify-center">
            <p className="text-sm text-muted-foreground">Navigation unavailable</p>
          </div>
        }
      >
        <Navigation />
      </ErrorBoundary>
      
      <main className="flex-1 overflow-auto">
        <div className="container mx-auto p-6">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </div>
      </main>
    </div>
  )
}