import React from 'react';
import { Card } from './ui/card';
import { Button } from './ui/button';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { 
  Maximize2, 
  Minimize2, 
  RotateCcw, 
  Sun, 
  Moon,
  TrendingUp
} from 'lucide-react';

interface TradingViewChartControlsProps {
  symbol: string;
  interval: string;
  theme: 'light' | 'dark';
  isFullscreen: boolean;
  onSymbolChange: (symbol: string) => void;
  onIntervalChange: (interval: string) => void;
  onThemeChange: (theme: 'light' | 'dark') => void;
  onToggleFullscreen: () => void;
  onReset: () => void;
  className?: string;
}

const TradingViewChartControls: React.FC<TradingViewChartControlsProps> = ({
  symbol,
  interval,
  theme,
  isFullscreen,
  onSymbolChange,
  onIntervalChange,
  onThemeChange,
  onToggleFullscreen,
  onReset,
  className = ''
}) => {
  const intervals = [
    { value: '1', label: '1m' },
    { value: '5', label: '5m' },
    { value: '15', label: '15m' },
    { value: '30', label: '30m' },
    { value: '60', label: '1h' },
    { value: '240', label: '4h' },
    { value: 'D', label: '1D' },
    { value: 'W', label: '1W' },
    { value: 'M', label: '1M' }
  ];

  const popularSymbols = [
    'AAPL', 'GOOGL', 'MSFT', 'AMZN', 'TSLA', 
    'META', 'NVDA', 'NFLX', 'AMD', 'INTC'
  ];

  const handleSymbolInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value.toUpperCase();
    onSymbolChange(value);
  };

  const handleSymbolInputKeyPress = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.currentTarget.blur();
    }
  };

  return (
    <Card className={`p-4 ${className}`}>
      <div className="flex flex-wrap items-center gap-4">
        {/* Symbol Input */}
        <div className="flex items-center space-x-2">
          <TrendingUp className="h-4 w-4 text-gray-500" />
          <div className="flex flex-col">
            <label className="text-xs text-gray-500 mb-1">Symbol</label>
            <input
              type="text"
              value={symbol}
              onChange={handleSymbolInputChange}
              onKeyPress={handleSymbolInputKeyPress}
              className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              placeholder="AAPL"
              maxLength={10}
            />
          </div>
        </div>

        {/* Popular Symbols */}
        <div className="flex flex-col">
          <label className="text-xs text-gray-500 mb-1">Quick Select</label>
          <div className="flex space-x-1">
            {popularSymbols.slice(0, 5).map((sym) => (
              <button
                key={sym}
                onClick={() => onSymbolChange(sym)}
                className={`px-2 py-1 text-xs rounded transition-colors ${
                  symbol === sym
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                {sym}
              </button>
            ))}
          </div>
        </div>

        {/* Interval Selector */}
        <div className="flex flex-col">
          <label className="text-xs text-gray-500 mb-1">Interval</label>
          <Select value={interval} onValueChange={onIntervalChange}>
            <SelectTrigger className="w-20">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {intervals.map((int) => (
                <SelectItem key={int.value} value={int.value}>
                  {int.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Theme Toggle */}
        <div className="flex flex-col">
          <label className="text-xs text-gray-500 mb-1">Theme</label>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onThemeChange(theme === 'light' ? 'dark' : 'light')}
            className="w-10 h-8"
          >
            {theme === 'light' ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </Button>
        </div>

        {/* Fullscreen Toggle */}
        <div className="flex flex-col">
          <label className="text-xs text-gray-500 mb-1">View</label>
          <Button
            variant="outline"
            size="sm"
            onClick={onToggleFullscreen}
            className="w-10 h-8"
          >
            {isFullscreen ? (
              <Minimize2 className="h-4 w-4" />
            ) : (
              <Maximize2 className="h-4 w-4" />
            )}
          </Button>
        </div>

        {/* Reset Button */}
        <div className="flex flex-col">
          <label className="text-xs text-gray-500 mb-1">Reset</label>
          <Button
            variant="outline"
            size="sm"
            onClick={onReset}
            className="w-10 h-8"
          >
            <RotateCcw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Keyboard Shortcuts Info */}
      <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
        <div className="text-xs text-gray-500 space-x-4">
          <span>F11: Fullscreen</span>
          <span>Ctrl+R: Reset</span>
          <span>Enter: Apply Symbol</span>
        </div>
      </div>
    </Card>
  );
};

export default TradingViewChartControls;