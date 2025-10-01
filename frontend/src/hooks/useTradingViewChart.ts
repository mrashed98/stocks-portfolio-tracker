import { useState, useEffect, useCallback } from 'react';

interface ChartConfig {
  symbol: string;
  interval: string;
  theme: 'light' | 'dark';
  height: number;
  autosize: boolean;
}

interface UseTradingViewChartProps {
  initialSymbol?: string;
  initialInterval?: string;
  initialTheme?: 'light' | 'dark';
}

export const useTradingViewChart = ({
  initialSymbol = 'AAPL',
  initialInterval = 'D',
  initialTheme = 'light'
}: UseTradingViewChartProps = {}) => {
  const [config, setConfig] = useState<ChartConfig>({
    symbol: initialSymbol,
    interval: initialInterval,
    theme: initialTheme,
    height: 400,
    autosize: true
  });

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Update symbol
  const updateSymbol = useCallback((symbol: string) => {
    setConfig(prev => ({ ...prev, symbol: symbol.toUpperCase() }));
  }, []);

  // Update interval
  const updateInterval = useCallback((interval: string) => {
    setConfig(prev => ({ ...prev, interval }));
  }, []);

  // Update theme
  const updateTheme = useCallback((theme: 'light' | 'dark') => {
    setConfig(prev => ({ ...prev, theme }));
  }, []);

  // Update height
  const updateHeight = useCallback((height: number) => {
    setConfig(prev => ({ ...prev, height }));
  }, []);

  // Toggle autosize
  const toggleAutosize = useCallback(() => {
    setConfig(prev => ({ ...prev, autosize: !prev.autosize }));
  }, []);

  // Toggle fullscreen
  const toggleFullscreen = useCallback(() => {
    setIsFullscreen(prev => !prev);
    // Adjust height based on fullscreen state
    setConfig(prev => ({
      ...prev,
      height: !isFullscreen ? window.innerHeight - 100 : 400
    }));
  }, [isFullscreen]);

  // Reset configuration
  const resetConfig = useCallback(() => {
    setConfig({
      symbol: initialSymbol,
      interval: initialInterval,
      theme: initialTheme,
      height: 400,
      autosize: true
    });
    setIsFullscreen(false);
    setError(null);
  }, [initialSymbol, initialInterval, initialTheme]);

  // Handle chart loading states
  const setChartLoading = useCallback((loading: boolean) => {
    setIsLoading(loading);
  }, []);

  const setChartError = useCallback((error: string | null) => {
    setError(error);
  }, []);

  // Listen for theme changes from system
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleThemeChange = (e: MediaQueryListEvent) => {
      // Auto theme handling would go here if we supported it
      console.log('System theme changed:', e.matches ? 'dark' : 'light');
    };

    mediaQuery.addEventListener('change', handleThemeChange);
    return () => mediaQuery.removeEventListener('change', handleThemeChange);
  }, [config.theme]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // F11 for fullscreen
      if (event.key === 'F11') {
        event.preventDefault();
        toggleFullscreen();
      }
      
      // Ctrl/Cmd + R for reset
      if ((event.ctrlKey || event.metaKey) && event.key === 'r') {
        event.preventDefault();
        resetConfig();
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [toggleFullscreen, resetConfig]);

  return {
    config,
    isFullscreen,
    isLoading,
    error,
    actions: {
      updateSymbol,
      updateInterval,
      updateTheme,
      updateHeight,
      toggleAutosize,
      toggleFullscreen,
      resetConfig,
      setChartLoading,
      setChartError
    }
  };
};

export default useTradingViewChart;