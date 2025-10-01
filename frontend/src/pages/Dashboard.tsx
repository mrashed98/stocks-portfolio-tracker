import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PortfolioDashboard } from '@/components/PortfolioDashboard'
import { PortfolioBuilder } from '@/components/PortfolioBuilder'
import { PortfolioDetail } from '@/components/PortfolioDetail'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import type { PortfolioResponse } from '@/types/portfolio'
import type { Portfolio } from '@/types'

type DashboardView = 'overview' | 'create-portfolio' | 'portfolio-detail'

export default function Dashboard() {
  const [currentView, setCurrentView] = useState<DashboardView>('overview')
  const [selectedPortfolio, setSelectedPortfolio] = useState<PortfolioResponse | null>(null)
  const navigate = useNavigate()

  const handleCreateNew = () => {
    setCurrentView('create-portfolio')
  }

  const handlePortfolioCreated = (portfolio: Portfolio) => {
    setCurrentView('overview')
    // Optionally show success message or navigate to the new portfolio
  }

  const handleViewPortfolio = (portfolio: PortfolioResponse) => {
    setSelectedPortfolio(portfolio)
    setCurrentView('portfolio-detail')
  }

  const handleBackToOverview = () => {
    setCurrentView('overview')
    setSelectedPortfolio(null)
  }

  const handleNavigateToStrategies = () => {
    navigate('/strategies')
  }

  return (
    <ErrorBoundary>
      <div className="space-y-6">
        {currentView === 'overview' && (
          <>
            <div>
              <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
              <p className="text-muted-foreground">
                Welcome to your portfolio management dashboard
              </p>
            </div>
            
            <PortfolioDashboard
              onCreateNew={handleCreateNew}
              onViewPortfolio={handleViewPortfolio}
            />
          </>
        )}

        {currentView === 'create-portfolio' && (
          <>
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold text-foreground">Create Portfolio</h1>
                <p className="text-muted-foreground">
                  Build a new portfolio using your investment strategies
                </p>
              </div>
              <button
                onClick={handleNavigateToStrategies}
                className="px-4 py-2 text-sm border border-input rounded-md hover:bg-accent"
              >
                Manage Strategies
              </button>
            </div>
            
            <PortfolioBuilder
              onPortfolioCreated={handlePortfolioCreated}
              onCancel={handleBackToOverview}
            />
          </>
        )}

        {currentView === 'portfolio-detail' && selectedPortfolio && (
          <>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <button
                  onClick={handleBackToOverview}
                  className="p-2 text-muted-foreground hover:text-foreground"
                >
                  <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                <div>
                  <h1 className="text-3xl font-bold text-foreground">{selectedPortfolio.name}</h1>
                  <p className="text-muted-foreground">
                    Portfolio details and performance tracking
                  </p>
                </div>
              </div>
            </div>
            
            <PortfolioDetail
              portfolioId={selectedPortfolio.id}
              onBack={handleBackToOverview}
            />
          </>
        )}
      </div>
    </ErrorBoundary>
  )
}