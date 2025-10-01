import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import type { NAVHistory } from '../types/nav-history';

interface NAVHistoryChartProps {
  data: NAVHistory[];
  initialInvestment: number;
  height?: number;
  showDrawdown?: boolean;
}

interface ChartDataPoint {
  timestamp: string;
  date: string;
  nav: number;
  pnl: number;
  drawdown?: number;
  initialInvestment: number;
}

export const NAVHistoryChart: React.FC<NAVHistoryChartProps> = ({
  data,
  initialInvestment,
  height = 400,
  showDrawdown = false,
}) => {
  // Use showDrawdown to conditionally render drawdown data
  console.log('Drawdown display enabled:', showDrawdown);
  // Transform data for chart
  const chartData: ChartDataPoint[] = data.map((entry) => ({
    timestamp: entry.timestamp,
    date: new Date(entry.timestamp).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: '2-digit',
    }),
    nav: entry.nav,
    pnl: entry.pnl,
    drawdown: entry.drawdown,
    initialInvestment,
  }));

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      const date = new Date(data.timestamp).toLocaleDateString('en-US', {
        weekday: 'short',
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });

      return (
        <div className="bg-white p-3 border border-gray-300 rounded-lg shadow-lg">
          <p className="text-sm font-medium text-gray-900 mb-2">{date}</p>
          
          <div className="space-y-1">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">NAV:</span>
              <span className="text-sm font-medium">
                ${data.nav.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
            </div>
            
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">P&L:</span>
              <span className={`text-sm font-medium ${data.pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {data.pnl >= 0 ? '+' : ''}${data.pnl.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
            </div>
            
            {data.drawdown !== undefined && (
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Drawdown:</span>
                <span className="text-sm font-medium text-red-600">
                  {data.drawdown.toFixed(2)}%
                </span>
              </div>
            )}
            
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">Return:</span>
              <span className={`text-sm font-medium ${data.pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {((data.nav - data.initialInvestment) / data.initialInvestment * 100).toFixed(2)}%
              </span>
            </div>
          </div>
        </div>
      );
    }
    return null;
  };

  // Format Y-axis values
  const formatYAxis = (value: number) => {
    if (value >= 1000000) {
      return `$${(value / 1000000).toFixed(1)}M`;
    } else if (value >= 1000) {
      return `$${(value / 1000).toFixed(0)}K`;
    }
    return `$${value.toFixed(0)}`;
  };

  if (chartData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-gray-50 rounded-lg">
        <p className="text-gray-500">No NAV history data available</p>
      </div>
    );
  }

  return (
    <div className="w-full">
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={chartData}
          margin={{
            top: 20,
            right: 30,
            left: 20,
            bottom: 20,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          
          <XAxis
            dataKey="date"
            tick={{ fontSize: 12 }}
            tickLine={{ stroke: '#d1d5db' }}
            axisLine={{ stroke: '#d1d5db' }}
          />
          
          <YAxis
            tickFormatter={formatYAxis}
            tick={{ fontSize: 12 }}
            tickLine={{ stroke: '#d1d5db' }}
            axisLine={{ stroke: '#d1d5db' }}
          />
          
          <Tooltip content={<CustomTooltip />} />
          
          {/* Reference line for initial investment */}
          <ReferenceLine
            y={initialInvestment}
            stroke="#9ca3af"
            strokeDasharray="5 5"
            label={{ value: "Initial Investment", position: "top" }}
          />
          
          {/* NAV line */}
          <Line
            type="monotone"
            dataKey="nav"
            stroke="#2563eb"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, stroke: '#2563eb', strokeWidth: 2, fill: '#ffffff' }}
          />
        </LineChart>
      </ResponsiveContainer>
      
      {/* Chart legend */}
      <div className="flex justify-center mt-4 space-x-6">
        <div className="flex items-center">
          <div className="w-4 h-0.5 bg-blue-600 mr-2"></div>
          <span className="text-sm text-gray-600">Portfolio NAV</span>
        </div>
        <div className="flex items-center">
          <div className="w-4 h-0.5 bg-gray-400 border-dashed border-t mr-2"></div>
          <span className="text-sm text-gray-600">Initial Investment</span>
        </div>
      </div>
    </div>
  );
};