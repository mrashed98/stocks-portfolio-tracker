import React, { useState, useEffect } from 'react';
import { Card } from './ui/card';
import { useToast } from '../hooks/use-toast';
import { portfolioService } from '../services/portfolioService';
import { PerformanceMetrics } from './PerformanceMetrics';
import { NAVHistoryChart } from './NAVHistoryChart';
import type { PortfolioResponse } from '../types/portfolio';
import type { PerformanceMetrics as PerformanceMetricsType, NAVHistory } from '../types/nav-history';

interface PortfolioDashboardProps {
  onCreateNew?: () => void;
  onViewPortfolio?: (portfolio: PortfolioResponse) => void;
}

interface PortfolioWithMetrics extends PortfolioResponse {
  metrics?: PerformanceMetricsType;
  history?: NAVHistory[];
}

export const PortfolioDashboard: React.FC<PortfolioDashboardProps> = ({
  onCreateNew,
  onViewPortfolio,
}) => {
  const { toast } = useToast();
  
  // State
  const [portfolios, setPortfolios] = useState<PortfolioWithMetrics[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdatingNAV, setIsUpdatingNAV] = useState<string | null>(null);
  const [selectedPortfolio, setSelectedPortfolio] = useState<PortfolioWithMetrics | null>(null);
  const [showPerformanceDetail, setShowPerformanceDetail] = useState(false);

  // Load portfolios on component mount
  useEffect(() => {
    loadPortfolios();
  }, []);

  const loadPortfolios = async () => {
    try {
      setIsLoading(true);
      const portfoliosData = await portfolioService.getPortfolios();
      
      // Load performance metrics and history for each portfolio
      const portfoliosWithMetrics = await Promise.all(
        portfoliosData.map(async (portfolio) => {
          try {
            const [metrics, history] = await Promise.all([
              portfolioService.getPortfolioPerformance(portfolio.id),
              portfolioService.getPortfolioHistory(portfolio.id),
            ]);
            return { ...portfolio, metrics, history };
          } catch (error) {
            console.error(`Failed to load metrics for portfolio ${portfolio.id}:`, error);
            return portfolio;
          }
        })
      );
      
      setPortfolios(portfoliosWithMetrics);
    } catch (error) {
      console.error('Failed to load portfolios:', error);
      toast({
        title: 'Error',
        description: 'Failed to load portfolios. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleUpdateNAV = async (portfolioId: string) => {
    try {
      setIsUpdatingNAV(portfolioId);
      await portfolioService.updateSinglePortfolioNAV(portfolioId);
      
      // Reload portfolios to get updated data
      await loadPortfolios();
      
      toast({
        title: 'Success',
        description: 'Portfolio NAV updated successfully.',
      });
    } catch (error: any) {
      console.error('Failed to update NAV:', error);
      const errorMessage = error.response?.data?.error || 'Failed to update NAV';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsUpdatingNAV(null);
    }
  };

  const handleViewPerformance = (portfolio: PortfolioWithMetrics) => {
    setSelectedPortfolio(portfolio);
    setShowPerformanceDetail(true);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (isLoading) {
    return (
      <Card className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading portfolios...</p>
          </div>
        </div>
      </Card>
    );
  }

  // Performance detail view
  if (showPerformanceDetail && selectedPortfolio) {
    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => setShowPerformanceDetail(false)}
              className="p-2 text-gray-400 hover:text-gray-600"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{selectedPortfolio.name} - Performance</h1>
              <p className="text-gray-600">Detailed performance analysis</p>
            </div>
          </div>
          <button
            onClick={() => onViewPortfolio?.(selectedPortfolio)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            View Full Details
          </button>
        </div>

        {/* Performance Metrics */}
        {selectedPortfolio.metrics && (
          <PerformanceMetrics
            metrics={selectedPortfolio.metrics}
            totalInvestment={selectedPortfolio.total_investment}
            currentNAV={selectedPortfolio.current_nav}
            lastUpdated={selectedPortfolio.history?.[selectedPortfolio.history.length - 1]?.timestamp}
          />
        )}

        {/* NAV History Chart */}
        {selectedPortfolio.history && selectedPortfolio.history.length > 0 && (
          <Card className="p-6">
            <h2 className="text-xl font-semibold mb-4">NAV History</h2>
            <NAVHistoryChart
              data={selectedPortfolio.history}
              initialInvestment={selectedPortfolio.total_investment}
              height={400}
            />
          </Card>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Portfolio Dashboard</h1>
        {onCreateNew && (
          <button
            onClick={onCreateNew}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Create New Portfolio
          </button>
        )}
      </div>

      {/* Portfolio Grid */}
      {portfolios.length === 0 ? (
        <Card className="p-8">
          <div className="text-center">
            <div className="mb-4">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No portfolios yet</h3>
            <p className="text-gray-600 mb-4">
              Create your first portfolio to start tracking your investments.
            </p>
            {onCreateNew && (
              <button
                onClick={onCreateNew}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Create Portfolio
              </button>
            )}
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {portfolios.map((portfolio) => (
            <Card key={portfolio.id} className="p-6 hover:shadow-lg transition-shadow">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{portfolio.name}</h3>
                  <p className="text-sm text-gray-600">
                    Created {formatDate(portfolio.created_at)}
                  </p>
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => handleUpdateNAV(portfolio.id)}
                    disabled={isUpdatingNAV === portfolio.id}
                    className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                    title="Update NAV"
                  >
                    {isUpdatingNAV === portfolio.id ? (
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
                    ) : (
                      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Portfolio Metrics */}
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Total Investment</span>
                  <span className="font-medium">{formatCurrency(portfolio.total_investment)}</span>
                </div>

                {portfolio.current_nav !== undefined && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Current Value</span>
                    <span className="font-medium">{formatCurrency(portfolio.current_nav)}</span>
                  </div>
                )}

                {portfolio.total_pnl !== undefined && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Total P&L</span>
                    <span className={`font-medium ${portfolio.total_pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {formatCurrency(portfolio.total_pnl)}
                    </span>
                  </div>
                )}

                {portfolio.metrics?.total_return_pct !== undefined && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Total Return</span>
                    <span className={`font-medium ${portfolio.metrics.total_return_pct >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {formatPercentage(portfolio.metrics.total_return_pct)}
                    </span>
                  </div>
                )}

                {portfolio.max_drawdown !== undefined && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Max Drawdown</span>
                    <span className="font-medium text-red-600">
                      {formatPercentage(portfolio.max_drawdown)}
                    </span>
                  </div>
                )}

                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Positions</span>
                  <span className="font-medium">{portfolio.positions?.length || 0}</span>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="mt-4 pt-4 border-t border-gray-200 space-y-2">
                <button
                  onClick={() => onViewPortfolio?.(portfolio)}
                  className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                >
                  View Details
                </button>
                <button
                  onClick={() => handleViewPerformance(portfolio)}
                  className="w-full px-4 py-2 border border-blue-600 text-blue-600 rounded-md hover:bg-blue-50 text-sm"
                >
                  View Performance
                </button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};