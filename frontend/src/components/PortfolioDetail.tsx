import React, { useState, useEffect } from 'react';
import { Card } from './ui/card';
import { useToast } from '../hooks/use-toast';
import { portfolioService } from '../services/portfolioService';
import { PerformanceMetrics } from './PerformanceMetrics';
import { NAVHistoryChart } from './NAVHistoryChart';
import { PositionPerformance } from './PositionPerformance';
import type { PortfolioResponse, AllocationPreview } from '../types/portfolio';
import type { NAVHistory, PerformanceMetrics as PerformanceMetricsType } from '../types/nav-history';

interface PortfolioDetailProps {
  portfolioId: string;
  onBack?: () => void;
  onEdit?: (portfolio: PortfolioResponse) => void;
}

export const PortfolioDetail: React.FC<PortfolioDetailProps> = ({
  portfolioId,
  onBack,
  onEdit,
}) => {
  const { toast } = useToast();
  
  // State
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [history, setHistory] = useState<NAVHistory[]>([]);
  const [metrics, setMetrics] = useState<PerformanceMetricsType | null>(null);
  const [rebalancePreview, setRebalancePreview] = useState<AllocationPreview | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdatingNAV, setIsUpdatingNAV] = useState(false);
  const [isGeneratingRebalance, setIsGeneratingRebalance] = useState(false);
  const [isRebalancing, setIsRebalancing] = useState(false);
  const [showRebalanceForm, setShowRebalanceForm] = useState(false);
  const [newInvestmentAmount, setNewInvestmentAmount] = useState<number>(0);

  // Load portfolio data on component mount
  useEffect(() => {
    loadPortfolioData();
  }, [portfolioId]);

  const loadPortfolioData = async () => {
    try {
      setIsLoading(true);
      
      // Load portfolio, history, and metrics in parallel
      const [portfolioData, historyData, metricsData] = await Promise.all([
        portfolioService.getPortfolio(portfolioId),
        portfolioService.getPortfolioHistory(portfolioId),
        portfolioService.getPortfolioPerformance(portfolioId),
      ]);
      
      setPortfolio(portfolioData);
      setHistory(historyData);
      setMetrics(metricsData);
      setNewInvestmentAmount(portfolioData.total_investment);
    } catch (error) {
      console.error('Failed to load portfolio data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load portfolio data. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleUpdateNAV = async () => {
    try {
      setIsUpdatingNAV(true);
      await portfolioService.updateSinglePortfolioNAV(portfolioId);
      
      // Reload portfolio data
      await loadPortfolioData();
      
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
      setIsUpdatingNAV(false);
    }
  };

  const handleGenerateRebalancePreview = async () => {
    if (!portfolio || newInvestmentAmount <= 0) return;

    try {
      setIsGeneratingRebalance(true);
      const preview = await portfolioService.generateRebalancePreview(portfolioId, newInvestmentAmount);
      setRebalancePreview(preview);
    } catch (error: any) {
      console.error('Failed to generate rebalance preview:', error);
      const errorMessage = error.response?.data?.error || 'Failed to generate rebalance preview';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsGeneratingRebalance(false);
    }
  };

  const handleRebalance = async () => {
    if (!rebalancePreview) return;

    try {
      setIsRebalancing(true);
      await portfolioService.rebalancePortfolio(portfolioId, newInvestmentAmount);
      
      // Reload portfolio data
      await loadPortfolioData();
      setRebalancePreview(null);
      setShowRebalanceForm(false);
      
      toast({
        title: 'Success',
        description: 'Portfolio rebalanced successfully.',
      });
    } catch (error: any) {
      console.error('Failed to rebalance portfolio:', error);
      const errorMessage = error.response?.data?.error || 'Failed to rebalance portfolio';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsRebalancing(false);
    }
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
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (isLoading) {
    return (
      <Card className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading portfolio...</p>
          </div>
        </div>
      </Card>
    );
  }

  if (!portfolio) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <p className="text-gray-600">Portfolio not found.</p>
          {onBack && (
            <button
              onClick={onBack}
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Go Back
            </button>
          )}
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div className="flex items-center space-x-4">
          {onBack && (
            <button
              onClick={onBack}
              className="p-2 text-gray-400 hover:text-gray-600"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
          )}
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{portfolio.name}</h1>
            <p className="text-gray-600">Created {formatDate(portfolio.created_at)}</p>
          </div>
        </div>
        <div className="flex space-x-3">
          <button
            onClick={handleUpdateNAV}
            disabled={isUpdatingNAV}
            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50 flex items-center"
          >
            {isUpdatingNAV && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600 mr-2"></div>
            )}
            Update NAV
          </button>
          <button
            onClick={() => setShowRebalanceForm(!showRebalanceForm)}
            className="px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700"
          >
            Rebalance
          </button>
          {onEdit && (
            <button
              onClick={() => onEdit(portfolio)}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Edit
            </button>
          )}
        </div>
      </div>

      {/* Performance Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <h3 className="text-sm font-medium text-gray-700">Total Investment</h3>
          <p className="text-2xl font-bold text-gray-900">{formatCurrency(portfolio.total_investment)}</p>
        </Card>
        
        {portfolio.current_nav !== undefined && (
          <Card className="p-4">
            <h3 className="text-sm font-medium text-gray-700">Current Value</h3>
            <p className="text-2xl font-bold text-gray-900">{formatCurrency(portfolio.current_nav)}</p>
          </Card>
        )}
        
        {portfolio.total_pnl !== undefined && (
          <Card className="p-4">
            <h3 className="text-sm font-medium text-gray-700">Total P&L</h3>
            <p className={`text-2xl font-bold ${portfolio.total_pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {formatCurrency(portfolio.total_pnl)}
            </p>
          </Card>
        )}
        
        {metrics?.total_return_pct !== undefined && (
          <Card className="p-4">
            <h3 className="text-sm font-medium text-gray-700">Total Return</h3>
            <p className={`text-2xl font-bold ${metrics.total_return_pct >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {formatPercentage(metrics.total_return_pct)}
            </p>
          </Card>
        )}
      </div>

      {/* Rebalance Form */}
      {showRebalanceForm && (
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Rebalance Portfolio</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                New Total Investment
              </label>
              <input
                type="number"
                value={newInvestmentAmount}
                onChange={(e) => setNewInvestmentAmount(Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="1"
                step="0.01"
              />
            </div>
            <div className="flex items-end">
              <button
                onClick={handleGenerateRebalancePreview}
                disabled={isGeneratingRebalance || newInvestmentAmount <= 0}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 flex items-center"
              >
                {isGeneratingRebalance && (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                )}
                Generate Preview
              </button>
            </div>
          </div>

          {rebalancePreview && (
            <div>
              <h3 className="text-lg font-medium mb-3">Rebalance Preview</h3>
              
              {/* Summary */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                <div className="bg-gray-50 p-3 rounded">
                  <p className="text-sm text-gray-600">New Investment</p>
                  <p className="text-lg font-semibold">{formatCurrency(rebalancePreview.total_investment)}</p>
                </div>
                <div className="bg-green-50 p-3 rounded">
                  <p className="text-sm text-gray-600">Total Allocated</p>
                  <p className="text-lg font-semibold text-green-600">{formatCurrency(rebalancePreview.total_allocated)}</p>
                </div>
                <div className="bg-yellow-50 p-3 rounded">
                  <p className="text-sm text-gray-600">Unallocated Cash</p>
                  <p className="text-lg font-semibold text-yellow-600">{formatCurrency(rebalancePreview.unallocated_cash)}</p>
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => {
                    setRebalancePreview(null);
                    setShowRebalanceForm(false);
                  }}
                  className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRebalance}
                  disabled={isRebalancing}
                  className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:bg-gray-400 flex items-center"
                >
                  {isRebalancing && (
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  )}
                  Confirm Rebalance
                </button>
              </div>
            </div>
          )}
        </Card>
      )}

      {/* Position Performance */}
      {portfolio.positions && (
        <PositionPerformance
          positions={portfolio.positions}
          lastUpdated={history.length > 0 ? history[history.length - 1].timestamp : undefined}
        />
      )}

      {/* Performance Metrics */}
      {metrics && (
        <PerformanceMetrics
          metrics={metrics}
          totalInvestment={portfolio.total_investment}
          currentNAV={portfolio.current_nav}
          lastUpdated={history.length > 0 ? history[history.length - 1].timestamp : undefined}
        />
      )}

      {/* NAV History Chart */}
      {history.length > 0 && (
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">NAV History Chart</h2>
          <NAVHistoryChart
            data={history}
            initialInvestment={portfolio.total_investment}
            height={400}
          />
        </Card>
      )}

      {/* NAV History */}
      {history.length > 0 && (
        <Card className="p-6">
          <h2 className="text-xl font-semibold mb-4">Recent NAV History</h2>
          
          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-gray-300">
              <thead>
                <tr className="bg-gray-50">
                  <th className="border border-gray-300 px-4 py-2 text-left">Date</th>
                  <th className="border border-gray-300 px-4 py-2 text-right">NAV</th>
                  <th className="border border-gray-300 px-4 py-2 text-right">P&L</th>
                  <th className="border border-gray-300 px-4 py-2 text-right">Drawdown</th>
                </tr>
              </thead>
              <tbody>
                {history.slice(-10).reverse().map((entry) => (
                  <tr key={entry.timestamp}>
                    <td className="border border-gray-300 px-4 py-2">
                      {formatDate(entry.timestamp)}
                    </td>
                    <td className="border border-gray-300 px-4 py-2 text-right">
                      {formatCurrency(entry.nav)}
                    </td>
                    <td className={`border border-gray-300 px-4 py-2 text-right ${
                      entry.pnl >= 0 ? 'text-green-600' : 'text-red-600'
                    }`}>
                      {formatCurrency(entry.pnl)}
                    </td>
                    <td className="border border-gray-300 px-4 py-2 text-right text-red-600">
                      {entry.drawdown !== undefined ? formatPercentage(entry.drawdown) : 'N/A'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}
    </div>
  );
};