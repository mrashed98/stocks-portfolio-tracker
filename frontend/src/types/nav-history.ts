import { z } from 'zod';
import type { Portfolio } from './portfolio';

// NAV History interfaces
export interface NAVHistory {
  portfolio_id: string;
  timestamp: string;
  nav: number;
  pnl: number;
  drawdown?: number;
  created_at: string;
  portfolio?: Portfolio;
}

export interface CreateNAVHistoryRequest {
  portfolio_id: string;
  timestamp: string;
  nav: number;
  pnl: number;
  drawdown?: number;
}

export interface NAVHistoryResponse {
  portfolio_id: string;
  timestamp: string;
  nav: number;
  pnl: number;
  drawdown?: number;
  created_at: string;
  portfolio?: Portfolio;
}

// Performance metrics interface
export interface PerformanceMetrics {
  total_return: number;
  total_return_pct: number;
  annualized_return?: number;
  max_drawdown?: number;
  current_drawdown?: number;
  volatility_pct?: number;
  sharpe_ratio?: number;
  days_active: number;
  high_water_mark: number;
}

// Chart data interfaces for visualization
export interface NAVChartData {
  timestamp: string;
  nav: number;
  pnl: number;
  drawdown?: number;
}

export interface PerformanceChartData {
  date: string;
  nav: number;
  benchmark?: number;
  drawdown: number;
}

// Zod validation schemas
export const createNAVHistorySchema = z.object({
  portfolio_id: z.string().uuid('Portfolio ID must be a valid UUID'),
  timestamp: z.string().datetime('Timestamp must be a valid ISO datetime string'),
  nav: z.number().min(0, 'NAV must be non-negative'),
  pnl: z.number(),
  drawdown: z.number().optional(),
});

// Helper functions for performance calculations
export const calculatePerformanceMetrics = (
  navHistory: NAVHistory[],
  initialInvestment: number
): PerformanceMetrics => {
  if (navHistory.length === 0) {
    return {
      total_return: 0,
      total_return_pct: 0,
      days_active: 0,
      high_water_mark: initialInvestment,
    };
  }

  const latest = navHistory[navHistory.length - 1];
  
  // Calculate total return
  const totalReturn = latest.nav - initialInvestment;
  const totalReturnPct = initialInvestment > 0 ? (totalReturn / initialInvestment) * 100 : 0;
  
  // Find high water mark and max drawdown
  let highWaterMark = initialInvestment;
  let maxDrawdown: number | undefined;
  
  for (const entry of navHistory) {
    if (entry.nav > highWaterMark) {
      highWaterMark = entry.nav;
    }
    
    if (entry.drawdown !== undefined) {
      if (maxDrawdown === undefined || entry.drawdown < maxDrawdown) {
        maxDrawdown = entry.drawdown;
      }
    }
  }
  
  // Calculate days active
  let daysActive = 0;
  if (navHistory.length > 1) {
    const firstEntry = new Date(navHistory[0].timestamp);
    const lastEntry = new Date(navHistory[navHistory.length - 1].timestamp);
    daysActive = Math.floor((lastEntry.getTime() - firstEntry.getTime()) / (1000 * 60 * 60 * 24));
  }
  
  // Calculate current drawdown
  let currentDrawdown: number | undefined;
  if (highWaterMark > 0 && latest.nav < highWaterMark) {
    currentDrawdown = ((latest.nav - highWaterMark) / highWaterMark) * 100;
  }
  
  // Calculate annualized return if we have enough data
  let annualizedReturn: number | undefined;
  if (daysActive > 0) {
    const years = daysActive / 365.25;
    if (years > 0 && initialInvestment > 0) {
      // Simplified annualized return calculation
      const finalRatio = latest.nav / initialInvestment;
      if (finalRatio > 0) {
        annualizedReturn = ((finalRatio - 1) / years) * 100;
      }
    }
  }
  
  return {
    total_return: totalReturn,
    total_return_pct: totalReturnPct,
    annualized_return: annualizedReturn,
    max_drawdown: maxDrawdown,
    current_drawdown: currentDrawdown,
    days_active: daysActive,
    high_water_mark: highWaterMark,
  };
};

// Helper function to format NAV history for charts
export const formatNAVHistoryForChart = (navHistory: NAVHistory[]): NAVChartData[] => {
  return navHistory.map(entry => ({
    timestamp: entry.timestamp,
    nav: entry.nav,
    pnl: entry.pnl,
    drawdown: entry.drawdown,
  }));
};

// Type inference from schemas
export type CreateNAVHistoryFormData = z.infer<typeof createNAVHistorySchema>;