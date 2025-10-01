import React, { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import TradingViewChart from '../components/TradingViewChart';
import TradingViewChartControls from '../components/TradingViewChartControls';
import { useTradingViewChart } from '../hooks/useTradingViewChart';
import { Card } from '../components/ui/card';

const ChartPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const symbolFromUrl = searchParams.get('symbol') || 'AAPL';
  const intervalFromUrl = searchParams.get('interval') || 'D';
  
  const { config, isFullscreen, isLoading, error, actions } = useTradingViewChart({
    initialSymbol: symbolFromUrl,
    initialInterval: intervalFromUrl,
    initialTheme: 'light'
  });

  // Update URL when symbol or interval changes
  useEffect(() => {
    const newParams = new URLSearchParams(searchParams);
    newParams.set('symbol', config.symbol);
    newParams.set('interval', config.interval);
    setSearchParams(newParams, { replace: true });
  }, [config.symbol, config.interval, searchParams, setSearchParams]);

  return (
    <div className="container mx-auto p-4 space-y-4">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Stock Charts
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Interactive TradingView charts with real-time data
          </p>
        </div>
        
        {/* Status Indicators */}
        <div className="flex items-center space-x-2">
          {isLoading && (
            <div className="flex items-center space-x-2 text-blue-600">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              <span className="text-sm">Loading...</span>
            </div>
          )}
          
          {error && (
            <div className="flex items-center space-x-2 text-red-600">
              <div className="h-2 w-2 bg-red-600 rounded-full"></div>
              <span className="text-sm">Chart Error</span>
            </div>
          )}
          
          {!isLoading && !error && (
            <div className="flex items-center space-x-2 text-green-600">
              <div className="h-2 w-2 bg-green-600 rounded-full"></div>
              <span className="text-sm">Live Data</span>
            </div>
          )}
        </div>
      </div>

      {/* Chart Controls */}
      {!isFullscreen && (
        <TradingViewChartControls
          symbol={config.symbol}
          interval={config.interval}
          theme={config.theme}
          isFullscreen={isFullscreen}
          onSymbolChange={actions.updateSymbol}
          onIntervalChange={actions.updateInterval}
          onThemeChange={actions.updateTheme}
          onToggleFullscreen={actions.toggleFullscreen}
          onReset={actions.resetConfig}
        />
      )}

      {/* Main Chart */}
      <div className={`${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900' : ''}`}>
        {isFullscreen && (
          <div className="p-4 border-b border-gray-200 dark:border-gray-700">
            <TradingViewChartControls
              symbol={config.symbol}
              interval={config.interval}
              theme={config.theme}
              isFullscreen={isFullscreen}
              onSymbolChange={actions.updateSymbol}
              onIntervalChange={actions.updateInterval}
              onThemeChange={actions.updateTheme}
              onToggleFullscreen={actions.toggleFullscreen}
              onReset={actions.resetConfig}
            />
          </div>
        )}
        
        <TradingViewChart
          symbol={config.symbol}
          interval={config.interval}
          theme={config.theme}
          height={config.height}
          autosize={config.autosize}
          className={isFullscreen ? 'h-full' : ''}
        />
      </div>

      {/* Additional Information */}
      {!isFullscreen && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Chart Info */}
          <Card className="p-4">
            <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">
              Chart Information
            </h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Symbol:</span>
                <span className="font-medium">{config.symbol}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Interval:</span>
                <span className="font-medium">{config.interval}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Theme:</span>
                <span className="font-medium capitalize">{config.theme}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Data Source:</span>
                <span className="font-medium">Portfolio API</span>
              </div>
            </div>
          </Card>

          {/* Features */}
          <Card className="p-4">
            <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">
              Features
            </h3>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-400">
              <li>• Real-time price data</li>
              <li>• Multiple timeframes</li>
              <li>• Interactive charts</li>
              <li>• Technical indicators</li>
              <li>• Fullscreen mode</li>
              <li>• Theme switching</li>
            </ul>
          </Card>

          {/* Keyboard Shortcuts */}
          <Card className="p-4">
            <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">
              Keyboard Shortcuts
            </h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Fullscreen:</span>
                <kbd className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs">F11</kbd>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Reset:</span>
                <kbd className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs">Ctrl+R</kbd>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Apply Symbol:</span>
                <kbd className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs">Enter</kbd>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Error Display */}
      {error && !isFullscreen && (
        <Card className="p-4 border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20">
          <div className="flex items-center space-x-2">
            <div className="h-4 w-4 bg-red-600 rounded-full flex-shrink-0"></div>
            <div>
              <h4 className="font-medium text-red-800 dark:text-red-200">
                Chart Error
              </h4>
              <p className="text-sm text-red-600 dark:text-red-300 mt-1">
                {error}
              </p>
              <button
                onClick={actions.resetConfig}
                className="mt-2 text-sm text-red-700 dark:text-red-200 underline hover:no-underline"
              >
                Try resetting the chart configuration
              </button>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
};

export default ChartPage;