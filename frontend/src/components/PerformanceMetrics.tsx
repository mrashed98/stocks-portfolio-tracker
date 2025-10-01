import React from 'react';
import { Card } from './ui/card';
import type { PerformanceMetrics as PerformanceMetricsType } from '../types/nav-history';

interface PerformanceMetricsProps {
  metrics: PerformanceMetricsType;
  totalInvestment: number;
  currentNAV?: number;
  isLoading?: boolean;
  lastUpdated?: string;
}

interface MetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  trend?: 'positive' | 'negative' | 'neutral';
  icon?: React.ReactNode;
  isLoading?: boolean;
}

const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  subtitle,
  trend = 'neutral',
  icon,
  isLoading = false,
}) => {
  const getTrendColor = () => {
    switch (trend) {
      case 'positive':
        return 'text-green-600';
      case 'negative':
        return 'text-red-600';
      default:
        return 'text-gray-900';
    }
  };

  if (isLoading) {
    return (
      <Card className="p-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-6 bg-gray-200 rounded w-1/2 mb-1"></div>
          <div className="h-3 bg-gray-200 rounded w-2/3"></div>
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className="text-sm font-medium text-gray-700 mb-1">{title}</h3>
          <p className={`text-2xl font-bold ${getTrendColor()}`}>
            {value}
          </p>
          {subtitle && (
            <p className="text-sm text-gray-500 mt-1">{subtitle}</p>
          )}
        </div>
        {icon && (
          <div className="ml-3 text-gray-400">
            {icon}
          </div>
        )}
      </div>
    </Card>
  );
};

export const PerformanceMetrics: React.FC<PerformanceMetricsProps> = ({
  metrics,
  totalInvestment,
  currentNAV,
  isLoading = false,
  lastUpdated,
}) => {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  };

  const formatPercentage = (value: number, showSign = true) => {
    const sign = showSign && value >= 0 ? '+' : '';
    return `${sign}${value.toFixed(2)}%`;
  };

  const formatNumber = (value: number) => {
    return value.toLocaleString('en-US');
  };

  const getTrend = (value: number): 'positive' | 'negative' | 'neutral' => {
    if (value > 0) return 'positive';
    if (value < 0) return 'negative';
    return 'neutral';
  };

  // Icons for different metrics
  const TrendUpIcon = () => (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
    </svg>
  );

  const TrendDownIcon = () => (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
    </svg>
  );

  const CalendarIcon = () => (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
    </svg>
  );

  const DollarIcon = () => (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
    </svg>
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold text-gray-900">Performance Metrics</h2>
        {lastUpdated && (
          <p className="text-sm text-gray-500">
            Last updated: {new Date(lastUpdated).toLocaleString('en-US', {
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </p>
        )}
      </div>

      {/* Key Performance Indicators */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Total Return"
          value={formatCurrency(metrics.total_return)}
          subtitle={formatPercentage(metrics.total_return_pct)}
          trend={getTrend(metrics.total_return)}
          icon={metrics.total_return >= 0 ? <TrendUpIcon /> : <TrendDownIcon />}
          isLoading={isLoading}
        />

        {currentNAV !== undefined && (
          <MetricCard
            title="Current Value"
            value={formatCurrency(currentNAV)}
            subtitle={`vs ${formatCurrency(totalInvestment)} invested`}
            trend={getTrend(currentNAV - totalInvestment)}
            icon={<DollarIcon />}
            isLoading={isLoading}
          />
        )}

        {metrics.annualized_return !== undefined && (
          <MetricCard
            title="Annualized Return"
            value={formatPercentage(metrics.annualized_return)}
            subtitle={`Over ${metrics.days_active} days`}
            trend={getTrend(metrics.annualized_return)}
            icon={<CalendarIcon />}
            isLoading={isLoading}
          />
        )}

        <MetricCard
          title="Days Active"
          value={formatNumber(metrics.days_active)}
          subtitle={`Since ${new Date(Date.now() - metrics.days_active * 24 * 60 * 60 * 1000).toLocaleDateString()}`}
          icon={<CalendarIcon />}
          isLoading={isLoading}
        />
      </div>

      {/* Risk Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {metrics.max_drawdown !== undefined && (
          <MetricCard
            title="Maximum Drawdown"
            value={formatPercentage(Math.abs(metrics.max_drawdown), false)}
            subtitle="Worst peak-to-trough decline"
            trend="negative"
            icon={<TrendDownIcon />}
            isLoading={isLoading}
          />
        )}

        {metrics.current_drawdown !== undefined && (
          <MetricCard
            title="Current Drawdown"
            value={formatPercentage(Math.abs(metrics.current_drawdown), false)}
            subtitle="Current decline from peak"
            trend={metrics.current_drawdown < 0 ? 'negative' : 'neutral'}
            icon={<TrendDownIcon />}
            isLoading={isLoading}
          />
        )}

        <MetricCard
          title="High Water Mark"
          value={formatCurrency(metrics.high_water_mark)}
          subtitle="Highest portfolio value"
          trend="positive"
          icon={<TrendUpIcon />}
          isLoading={isLoading}
        />
      </div>

      {/* Advanced Metrics (if available) */}
      {(metrics.volatility_pct !== undefined || metrics.sharpe_ratio !== undefined) && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {metrics.volatility_pct !== undefined && (
            <MetricCard
              title="Volatility"
              value={formatPercentage(metrics.volatility_pct, false)}
              subtitle="Standard deviation of returns"
              isLoading={isLoading}
            />
          )}

          {metrics.sharpe_ratio !== undefined && (
            <MetricCard
              title="Sharpe Ratio"
              value={metrics.sharpe_ratio.toFixed(2)}
              subtitle="Risk-adjusted return"
              trend={metrics.sharpe_ratio > 1 ? 'positive' : metrics.sharpe_ratio < 0 ? 'negative' : 'neutral'}
              isLoading={isLoading}
            />
          )}
        </div>
      )}

      {/* Data Staleness Indicator */}
      {lastUpdated && (
        <div className="mt-4 p-3 bg-gray-50 rounded-lg">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="w-2 h-2 bg-green-400 rounded-full mr-2"></div>
              <span className="text-sm text-gray-600">Market data is current</span>
            </div>
            <span className="text-xs text-gray-500">
              Updated {new Date(lastUpdated).toLocaleTimeString()}
            </span>
          </div>
        </div>
      )}
    </div>
  );
};