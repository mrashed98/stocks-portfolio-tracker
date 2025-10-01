import { useState } from 'react'
import { PortfolioDashboard } from '@/components/PortfolioDashboard'
import { PortfolioBuilder } from '@/components/PortfolioBuilder'
import { PortfolioDetail } from '@/components/PortfolioDetail'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import type { PortfolioResponse } from '@/types/portfolio'
import type { Portfolio } from '@/types'

type PortfolioView = 'list' | 'create' | 'detail'

export default function Portfolios() {
  const [currentView, setCurrentView] = useState<PortfolioView>('list')
  const [selectedPortfolio, setSelectedPortfolio] = useState<PortfolioResponse | null>(null)

  const handleCreateNew = () => {
    setCurrentView('create')
  }

  const handlePortfolioCreated = (portfolio: Portfolio) => {
    setCurrentView('list')
  }

  const handleViewPortfolio = (portfolio: PortfolioResponse) => {
    setSelectedPortfolio(portfolio)
    setCurrentView('detail')
  }

  const handleBackToList = () => {
    setCurrentView('list')
    setSelectedPortfolio(null)
  }

  return (
    <ErrorBoundary>
      <div className="space-y-6">
        {currentView === 'list' && (
          <>
            <div>
              <h1 className="text-3xl font-bold text-foreground">Portfolios</h1>
              <p className="text-muted-foreground">
                Manage and track your investment portfolios
              </p>
            </div>
            
            <PortfolioDashboard
              onCreateNew={handleCreateNew}
              onViewPortfolio={handleViewPortfolio}
            />
          </>
        )}

        {currentView === 'create' && (
          <>
            <div>
              <h1 className="text-3xl font-bold text-foreground">Create Portfolio</h1>
              <p className="text-muted-foreground">
                Build a new portfolio using your investment strategies
              </p>
            </div>
            
            <PortfolioBuilder
              onPortfolioCreated={handlePortfolioCreated}
              onCancel={handleBackToList}
            />
          </>
        )}

        {currentView === 'detail' && selectedPortfolio && (
          <>
            <div className="flex items-center space-x-4">
              <button
                onClick={handleBackToList}
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
            
            <PortfolioDetail
              portfolioId={selectedPortfolio.id}
              onBack={handleBackToList}
            />
          </>
        )}
      </div>
    </ErrorBoundary>
  )
}