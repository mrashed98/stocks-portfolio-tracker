import React from 'react';
import TradingViewChart from './TradingViewChart';
import { Card } from './ui/card';

interface StockChartProps {
  symbol: string;
  height?: number;
  showControls?: boolean;
  className?: string;
}

/**
 * Simple stock chart component for embedding in other pages
 * This is a simplified version of the full TradingView chart
 */
const StockChart: React.FC<StockChartProps> = ({
  symbol,
  height = 300,
  showControls = false,
  className = ''
}) => {
  if (!symbol) {
    return (
      <Card className={`p-6 ${className}`} style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-gray-500 dark:text-gray-400">
              No symbol provided for chart
            </p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <div className={className}>
      {showControls && (
        <div className="mb-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                {symbol}
              </span>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                Live Chart
              </span>
            </div>
            <a
              href={`/charts?symbol=${symbol}`}
              className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
            >
              View Full Chart →
            </a>
          </div>
        </div>
      )}
      
      <TradingViewChart
        symbol={symbol}
        interval="D"
        theme="light"
        height={height}
        autosize={true}
      />
    </div>
  );
};

export default StockChart;