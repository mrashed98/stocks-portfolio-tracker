import React, { useState, useEffect, useCallback } from 'react';
import { Card } from './ui/card';
import { useToast } from '../hooks/use-toast';
import { portfolioService } from '../services/portfolioService';
import { strategyService } from '../services/strategyService';
import type { 
  Strategy,
  AllocationPreview,
  AllocationRequest,
  AllocationConstraints,
  CreatePortfolioRequest,
  Portfolio
} from '../types';

interface PortfolioBuilderProps {
  onPortfolioCreated?: (portfolio: Portfolio) => void;
  onCancel?: () => void;
}

interface FormData {
  portfolioName: string;
  totalInvestment: number;
  selectedStrategies: string[];
  constraints: AllocationConstraints;
  excludedStocks: string[];
}

const DEFAULT_CONSTRAINTS: AllocationConstraints = {
  max_allocation_per_stock: 20, // 20%
  min_allocation_amount: 100,   // $100
};

export const PortfolioBuilder: React.FC<PortfolioBuilderProps> = ({
  onPortfolioCreated,
  onCancel,
}) => {
  const { toast } = useToast();
  
  // State
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [formData, setFormData] = useState<FormData>({
    portfolioName: '',
    totalInvestment: 10000,
    selectedStrategies: [],
    constraints: DEFAULT_CONSTRAINTS,
    excludedStocks: [],
  });
  const [allocationPreview, setAllocationPreview] = useState<AllocationPreview | null>(null);
  const [isLoadingStrategies, setIsLoadingStrategies] = useState(true);
  const [isGeneratingPreview, setIsGeneratingPreview] = useState(false);
  const [isCreatingPortfolio, setIsCreatingPortfolio] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  // Load strategies on component mount
  useEffect(() => {
    loadStrategies();
  }, []);

  // Generate preview when form data changes
  useEffect(() => {
    if (formData.selectedStrategies.length > 0 && formData.totalInvestment > 0) {
      generatePreview();
    } else {
      setAllocationPreview(null);
      setPreviewError(null);
    }
  }, [formData.selectedStrategies, formData.totalInvestment, formData.constraints, formData.excludedStocks]);

  const loadStrategies = async () => {
    try {
      setIsLoadingStrategies(true);
      const strategiesData = await strategyService.getStrategies();
      setStrategies(strategiesData);
    } catch (error) {
      console.error('Failed to load strategies:', error);
      toast({
        title: 'Error',
        description: 'Failed to load strategies. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setIsLoadingStrategies(false);
    }
  };

  const generatePreview = useCallback(async () => {
    if (formData.selectedStrategies.length === 0 || formData.totalInvestment <= 0) {
      return;
    }

    try {
      setIsGeneratingPreview(true);
      setPreviewError(null);

      const allocationRequest: AllocationRequest = {
        strategy_ids: formData.selectedStrategies,
        total_investment: formData.totalInvestment,
        constraints: formData.constraints,
        excluded_stocks: formData.excludedStocks,
      };

      let preview: AllocationPreview;
      if (formData.excludedStocks.length > 0) {
        preview = await portfolioService.generateAllocationPreviewWithExclusions(
          allocationRequest,
          formData.excludedStocks
        );
      } else {
        preview = await portfolioService.generateAllocationPreview(allocationRequest);
      }

      setAllocationPreview(preview);
    } catch (error: any) {
      console.error('Failed to generate allocation preview:', error);
      const errorMessage = error.response?.data?.error || 'Failed to generate allocation preview';
      setPreviewError(errorMessage);
      setAllocationPreview(null);
    } finally {
      setIsGeneratingPreview(false);
    }
  }, [formData]);

  const handleStrategyToggle = (strategyId: string) => {
    setFormData(prev => ({
      ...prev,
      selectedStrategies: prev.selectedStrategies.includes(strategyId)
        ? prev.selectedStrategies.filter(id => id !== strategyId)
        : [...prev.selectedStrategies, strategyId],
    }));
  };

  const handleStockExclusion = (stockId: string) => {
    setFormData(prev => ({
      ...prev,
      excludedStocks: prev.excludedStocks.includes(stockId)
        ? prev.excludedStocks.filter(id => id !== stockId)
        : [...prev.excludedStocks, stockId],
    }));
  };

  const handleConstraintChange = (field: keyof AllocationConstraints, value: number) => {
    setFormData(prev => ({
      ...prev,
      constraints: {
        ...prev.constraints,
        [field]: value,
      },
    }));
  };

  const handleCreatePortfolio = async () => {
    if (!allocationPreview || !formData.portfolioName.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Please provide a portfolio name and generate a valid allocation preview.',
        variant: 'destructive',
      });
      return;
    }

    try {
      setIsCreatingPortfolio(true);

      // Convert allocation preview to positions
      const positions = allocationPreview.allocations.map(allocation => ({
        stock_id: allocation.stock_id,
        quantity: allocation.quantity,
        entry_price: allocation.price,
        allocation_value: allocation.actual_value,
        strategy_contrib: allocation.strategy_contrib,
      }));

      const createRequest: CreatePortfolioRequest = {
        name: formData.portfolioName.trim(),
        total_investment: formData.totalInvestment,
        positions,
      };

      const portfolio = await portfolioService.createPortfolio(createRequest);

      toast({
        title: 'Success',
        description: `Portfolio "${portfolio.name}" created successfully!`,
      });

      if (onPortfolioCreated) {
        onPortfolioCreated(portfolio);
      }
    } catch (error: any) {
      console.error('Failed to create portfolio:', error);
      const errorMessage = error.response?.data?.error || 'Failed to create portfolio';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsCreatingPortfolio(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${value.toFixed(2)}%`;
  };

  if (isLoadingStrategies) {
    return (
      <Card className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading strategies...</p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Portfolio Configuration */}
      <Card className="p-6">
        <h2 className="text-xl font-semibold mb-4">Portfolio Configuration</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Portfolio Name
            </label>
            <input
              type="text"
              value={formData.portfolioName}
              onChange={(e) => setFormData(prev => ({ ...prev, portfolioName: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter portfolio name"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Total Investment
            </label>
            <input
              type="number"
              value={formData.totalInvestment}
              onChange={(e) => setFormData(prev => ({ ...prev, totalInvestment: Number(e.target.value) }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter total investment amount"
              min="1"
              step="0.01"
            />
          </div>
        </div>

        {/* Strategy Selection */}
        <div className="mb-6">
          <h3 className="text-lg font-medium mb-3">Select Strategies</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {strategies.map((strategy) => (
              <div
                key={strategy.id}
                className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                  formData.selectedStrategies.includes(strategy.id)
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
                onClick={() => handleStrategyToggle(strategy.id)}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">{strategy.name}</h4>
                    <p className="text-sm text-gray-600">
                      {strategy.weight_mode === 'percent' ? `${strategy.weight_value}%` : formatCurrency(strategy.weight_value)}
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={formData.selectedStrategies.includes(strategy.id)}
                    onChange={() => handleStrategyToggle(strategy.id)}
                    className="h-4 w-4 text-blue-600"
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Allocation Constraints */}
        <div>
          <h3 className="text-lg font-medium mb-3">Allocation Constraints</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Max Allocation per Stock (%)
              </label>
              <input
                type="number"
                value={formData.constraints.max_allocation_per_stock}
                onChange={(e) => handleConstraintChange('max_allocation_per_stock', Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="1"
                max="100"
                step="0.1"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Min Allocation Amount ($)
              </label>
              <input
                type="number"
                value={formData.constraints.min_allocation_amount}
                onChange={(e) => handleConstraintChange('min_allocation_amount', Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="0"
                step="0.01"
              />
            </div>
          </div>
        </div>
      </Card>

      {/* Allocation Preview */}
      <Card className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Allocation Preview</h2>
          {isGeneratingPreview && (
            <div className="flex items-center text-blue-600">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
              Calculating...
            </div>
          )}
        </div>

        {previewError && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
            <p className="text-red-800">{previewError}</p>
          </div>
        )}

        {allocationPreview && (
          <div>
            {/* Summary */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
              <div className="bg-gray-50 p-4 rounded-lg">
                <h4 className="font-medium text-gray-700">Total Investment</h4>
                <p className="text-2xl font-bold text-gray-900">
                  {formatCurrency(allocationPreview.total_investment)}
                </p>
              </div>
              <div className="bg-green-50 p-4 rounded-lg">
                <h4 className="font-medium text-gray-700">Total Allocated</h4>
                <p className="text-2xl font-bold text-green-600">
                  {formatCurrency(allocationPreview.total_allocated)}
                </p>
              </div>
              <div className="bg-yellow-50 p-4 rounded-lg">
                <h4 className="font-medium text-gray-700">Unallocated Cash</h4>
                <p className="text-2xl font-bold text-yellow-600">
                  {formatCurrency(allocationPreview.unallocated_cash)}
                </p>
              </div>
            </div>

            {/* Allocations Table */}
            <div className="overflow-x-auto">
              <table className="w-full border-collapse border border-gray-300">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="border border-gray-300 px-4 py-2 text-left">Stock</th>
                    <th className="border border-gray-300 px-4 py-2 text-right">Weight</th>
                    <th className="border border-gray-300 px-4 py-2 text-right">Price</th>
                    <th className="border border-gray-300 px-4 py-2 text-right">Quantity</th>
                    <th className="border border-gray-300 px-4 py-2 text-right">Value</th>
                    <th className="border border-gray-300 px-4 py-2 text-center">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {allocationPreview.allocations.map((allocation) => (
                    <tr
                      key={allocation.stock_id}
                      className={formData.excludedStocks.includes(allocation.stock_id) ? 'bg-red-50' : ''}
                    >
                      <td className="border border-gray-300 px-4 py-2">
                        <div>
                          <div className="font-medium">{allocation.ticker}</div>
                          <div className="text-sm text-gray-600">{allocation.name}</div>
                        </div>
                      </td>
                      <td className="border border-gray-300 px-4 py-2 text-right">
                        {formatPercentage(allocation.weight)}
                      </td>
                      <td className="border border-gray-300 px-4 py-2 text-right">
                        {formatCurrency(allocation.price)}
                      </td>
                      <td className="border border-gray-300 px-4 py-2 text-right">
                        {allocation.quantity}
                      </td>
                      <td className="border border-gray-300 px-4 py-2 text-right">
                        {formatCurrency(allocation.actual_value)}
                      </td>
                      <td className="border border-gray-300 px-4 py-2 text-center">
                        <button
                          onClick={() => handleStockExclusion(allocation.stock_id)}
                          className={`px-3 py-1 rounded text-sm ${
                            formData.excludedStocks.includes(allocation.stock_id)
                              ? 'bg-green-100 text-green-700 hover:bg-green-200'
                              : 'bg-red-100 text-red-700 hover:bg-red-200'
                          }`}
                        >
                          {formData.excludedStocks.includes(allocation.stock_id) ? 'Restore' : 'Exclude'}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {!allocationPreview && !previewError && formData.selectedStrategies.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            Select at least one strategy to generate allocation preview
          </div>
        )}
      </Card>

      {/* Action Buttons */}
      <div className="flex justify-end space-x-4">
        {onCancel && (
          <button
            onClick={onCancel}
            className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
            disabled={isCreatingPortfolio}
          >
            Cancel
          </button>
        )}
        <button
          onClick={handleCreatePortfolio}
          disabled={!allocationPreview || !formData.portfolioName.trim() || isCreatingPortfolio}
          className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center"
        >
          {isCreatingPortfolio && (
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
          )}
          Create Portfolio
        </button>
      </div>
    </div>
  );
};