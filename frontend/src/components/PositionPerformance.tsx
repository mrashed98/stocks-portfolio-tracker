import React, { useState } from 'react';
import { Card } from './ui/card';
import type { Position } from '../types/position';

interface PositionPerformanceProps {
  positions: Position[];
  isLoading?: boolean;
  lastUpdated?: string;
}

interface SortConfig {
  key: keyof Position | 'pnl_percentage' | 'current_value';
  direction: 'asc' | 'desc';
}

export const PositionPerformance: React.FC<PositionPerformanceProps> = ({
  positions,
  isLoading = false,
  lastUpdated,
}) => {
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: 'allocation_value', direction: 'desc' });

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    const sign = value >= 0 ? '+' : '';
    return `${sign}${value.toFixed(2)}%`;
  };

  const handleSort = (key: keyof Position | 'pnl_percentage' | 'current_value') => {
    let direction: 'asc' | 'desc' = 'desc';
    if (sortConfig.key === key && sortConfig.direction === 'desc') {
      direction = 'asc';
    }
    setSortConfig({ key, direction });
  };

  const getSortedPositions = () => {
    const sortedPositions = [...positions];
    
    sortedPositions.sort((a, b) => {
      let aValue: number;
      let bValue: number;

      switch (sortConfig.key) {
        case 'pnl_percentage':
          aValue = a.pnl_percentage || 0;
          bValue = b.pnl_percentage || 0;
          break;
        case 'current_value':
          aValue = a.current_value || a.allocation_value;
          bValue = b.current_value || b.allocation_value;
          break;
        case 'allocation_value':
          aValue = a.allocation_value;
          bValue = b.allocation_value;
          break;
        case 'quantity':
          aValue = a.quantity;
          bValue = b.quantity;
          break;
        case 'entry_price':
          aValue = a.entry_price;
          bValue = b.entry_price;
          break;
        default:
          return 0;
      }

      if (sortConfig.direction === 'asc') {
        return aValue - bValue;
      }
      return bValue - aValue;
    });

    return sortedPositions;
  };

  const getSortIcon = (columnKey: keyof Position | 'pnl_percentage' | 'current_value') => {
    if (sortConfig.key !== columnKey) {
      return (
        <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
        </svg>
      );
    }

    return sortConfig.direction === 'asc' ? (
      <svg className="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
      </svg>
    ) : (
      <svg className="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    );
  };

  const calculateTotals = () => {
    const totalAllocation = positions.reduce((sum, pos) => sum + pos.allocation_value, 0);
    const totalCurrentValue = positions.reduce((sum, pos) => sum + (pos.current_value || pos.allocation_value), 0);
    const totalPnL = positions.reduce((sum, pos) => sum + (pos.pnl || 0), 0);
    const totalPnLPercentage = totalAllocation > 0 ? (totalPnL / totalAllocation) * 100 : 0;

    return {
      totalAllocation,
      totalCurrentValue,
      totalPnL,
      totalPnLPercentage,
    };
  };

  const totals = calculateTotals();

  if (isLoading) {
    return (
      <Card className="p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-12 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </Card>
    );
  }

  if (positions.length === 0) {
    return (
      <Card className="p-6">
        <div className="text-center py-8">
          <svg
            className="mx-auto h-12 w-12 text-gray-400 mb-4"
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
          <h3 className="text-lg font-medium text-gray-900 mb-2">No positions</h3>
          <p className="text-gray-600">This portfolio doesn't have any positions yet.</p>
        </div>
      </Card>
    );
  }

  const sortedPositions = getSortedPositions();

  return (
    <Card className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-semibold text-gray-900">Position Performance</h2>
        {lastUpdated && (
          <div className="flex items-center text-sm text-gray-500">
            <div className="w-2 h-2 bg-green-400 rounded-full mr-2"></div>
            Last updated: {new Date(lastUpdated).toLocaleTimeString()}
          </div>
        )}
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-50 p-4 rounded-lg">
          <h3 className="text-sm font-medium text-gray-700">Total Positions</h3>
          <p className="text-2xl font-bold text-gray-900">{positions.length}</p>
        </div>
        
        <div className="bg-blue-50 p-4 rounded-lg">
          <h3 className="text-sm font-medium text-gray-700">Total Allocation</h3>
          <p className="text-2xl font-bold text-blue-900">{formatCurrency(totals.totalAllocation)}</p>
        </div>
        
        <div className="bg-green-50 p-4 rounded-lg">
          <h3 className="text-sm font-medium text-gray-700">Current Value</h3>
          <p className="text-2xl font-bold text-green-900">{formatCurrency(totals.totalCurrentValue)}</p>
        </div>
        
        <div className={`p-4 rounded-lg ${totals.totalPnL >= 0 ? 'bg-green-50' : 'bg-red-50'}`}>
          <h3 className="text-sm font-medium text-gray-700">Total P&L</h3>
          <p className={`text-2xl font-bold ${totals.totalPnL >= 0 ? 'text-green-900' : 'text-red-900'}`}>
            {formatCurrency(totals.totalPnL)}
          </p>
          <p className={`text-sm ${totals.totalPnL >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            {formatPercentage(totals.totalPnLPercentage)}
          </p>
        </div>
      </div>

      {/* Positions Table */}
      <div className="overflow-x-auto">
        <table className="w-full border-collapse">
          <thead>
            <tr className="border-b border-gray-200">
              <th className="text-left py-3 px-4 font-medium text-gray-700">Stock</th>
              
              <th 
                className="text-right py-3 px-4 font-medium text-gray-700 cursor-pointer hover:bg-gray-50"
                onClick={() => handleSort('quantity')}
              >
                <div className="flex items-center justify-end">
                  Quantity
                  {getSortIcon('quantity')}
                </div>
              </th>
              
              <th 
                className="text-right py-3 px-4 font-medium text-gray-700 cursor-pointer hover:bg-gray-50"
                onClick={() => handleSort('entry_price')}
              >
                <div className="flex items-center justify-end">
                  Entry Price
                  {getSortIcon('entry_price')}
                </div>
              </th>
              
              <th className="text-right py-3 px-4 font-medium text-gray-700">Current Price</th>
              
              <th 
                className="text-right py-3 px-4 font-medium text-gray-700 cursor-pointer hover:bg-gray-50"
                onClick={() => handleSort('allocation_value')}
              >
                <div className="flex items-center justify-end">
                  Initial Value
                  {getSortIcon('allocation_value')}
                </div>
              </th>
              
              <th 
                className="text-right py-3 px-4 font-medium text-gray-700 cursor-pointer hover:bg-gray-50"
                onClick={() => handleSort('current_value')}
              >
                <div className="flex items-center justify-end">
                  Current Value
                  {getSortIcon('current_value')}
                </div>
              </th>
              
              <th className="text-right py-3 px-4 font-medium text-gray-700">P&L</th>
              
              <th 
                className="text-right py-3 px-4 font-medium text-gray-700 cursor-pointer hover:bg-gray-50"
                onClick={() => handleSort('pnl_percentage')}
              >
                <div className="flex items-center justify-end">
                  P&L %
                  {getSortIcon('pnl_percentage')}
                </div>
              </th>
            </tr>
          </thead>
          
          <tbody>
            {sortedPositions.map((position) => (
              <tr key={position.stock_id} className="border-b border-gray-100 hover:bg-gray-50">
                <td className="py-3 px-4">
                  <div>
                    <div className="font-medium text-gray-900">
                      {position.stock?.ticker || 'N/A'}
                    </div>
                    <div className="text-sm text-gray-600 truncate max-w-32">
                      {position.stock?.name || 'Unknown'}
                    </div>
                  </div>
                </td>
                
                <td className="py-3 px-4 text-right font-medium">
                  {position.quantity.toLocaleString()}
                </td>
                
                <td className="py-3 px-4 text-right">
                  {formatCurrency(position.entry_price)}
                </td>
                
                <td className="py-3 px-4 text-right">
                  {position.current_price ? (
                    <div>
                      <div>{formatCurrency(position.current_price)}</div>
                      {position.current_price !== position.entry_price && (
                        <div className={`text-xs ${
                          position.current_price > position.entry_price ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {formatPercentage(((position.current_price - position.entry_price) / position.entry_price) * 100)}
                        </div>
                      )}
                    </div>
                  ) : (
                    <span className="text-gray-400">N/A</span>
                  )}
                </td>
                
                <td className="py-3 px-4 text-right">
                  {formatCurrency(position.allocation_value)}
                </td>
                
                <td className="py-3 px-4 text-right font-medium">
                  {position.current_value ? (
                    formatCurrency(position.current_value)
                  ) : (
                    <span className="text-gray-400">{formatCurrency(position.allocation_value)}</span>
                  )}
                </td>
                
                <td className={`py-3 px-4 text-right font-medium ${
                  position.pnl !== undefined 
                    ? position.pnl >= 0 ? 'text-green-600' : 'text-red-600'
                    : 'text-gray-400'
                }`}>
                  {position.pnl !== undefined ? formatCurrency(position.pnl) : 'N/A'}
                </td>
                
                <td className={`py-3 px-4 text-right font-medium ${
                  position.pnl_percentage !== undefined 
                    ? position.pnl_percentage >= 0 ? 'text-green-600' : 'text-red-600'
                    : 'text-gray-400'
                }`}>
                  {position.pnl_percentage !== undefined ? formatPercentage(position.pnl_percentage) : 'N/A'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Data staleness indicator */}
      <div className="mt-4 text-xs text-gray-500 text-center">
        {lastUpdated ? (
          `Market data last updated: ${new Date(lastUpdated).toLocaleString()}`
        ) : (
          'Market data may not be current'
        )}
      </div>
    </Card>
  );
};