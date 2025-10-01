import { z } from 'zod';
import type { Position } from './position';
import type { NAVHistory } from './nav-history';

// Portfolio interfaces
export interface Portfolio {
  id: string;
  user_id: string;
  name: string;
  total_investment: number;
  created_at: string;
  updated_at: string;
  positions?: Position[];
  nav_history?: NAVHistory[];
}

export interface CreatePortfolioRequest {
  name: string;
  total_investment: number;
  positions: CreatePositionRequest[];
}

export interface UpdatePortfolioRequest {
  name?: string;
  total_investment?: number;
}

export interface PortfolioResponse {
  id: string;
  user_id: string;
  name: string;
  total_investment: number;
  created_at: string;
  updated_at: string;
  positions?: Position[];
  nav_history?: NAVHistory[];
  current_nav?: number;
  total_pnl?: number;
  max_drawdown?: number;
}

// Allocation interfaces
export interface AllocationPreview {
  total_investment: number;
  allocations: StockAllocation[];
  unallocated_cash: number;
  total_allocated: number;
  constraints: AllocationConstraints;
}

export interface StockAllocation {
  stock_id: string;
  ticker: string;
  name: string;
  weight: number;
  allocation_value: number;
  price: number;
  quantity: number;
  actual_value: number;
  strategy_contrib: Record<string, number>;
}

export interface AllocationConstraints {
  max_allocation_per_stock: number;
  min_allocation_amount: number;
}

export interface AllocationRequest {
  strategy_ids: string[];
  total_investment: number;
  constraints: AllocationConstraints;
  excluded_stocks?: string[];
}

// Import CreatePositionRequest type
import type { CreatePositionRequest } from './position';

// Zod validation schemas
export const allocationConstraintsSchema = z.object({
  max_allocation_per_stock: z.number()
    .positive('Max allocation per stock must be greater than 0')
    .max(100, 'Max allocation per stock cannot exceed 100%'),
  min_allocation_amount: z.number()
    .min(0, 'Min allocation amount must be non-negative'),
});

export const allocationRequestSchema = z.object({
  strategy_ids: z.array(z.string().uuid('Strategy ID must be a valid UUID'))
    .min(1, 'At least one strategy must be selected'),
  total_investment: z.number().positive('Total investment must be greater than 0'),
  constraints: allocationConstraintsSchema,
  excluded_stocks: z.array(z.string().uuid('Stock ID must be a valid UUID')).optional(),
});

export const createPortfolioSchema = z.object({
  name: z.string().min(1, 'Portfolio name is required').max(255, 'Portfolio name must be at most 255 characters'),
  total_investment: z.number().positive('Total investment must be greater than 0'),
  positions: z.array(z.object({
    stock_id: z.string().uuid('Stock ID must be a valid UUID'),
    quantity: z.number().int().positive('Quantity must be a positive integer'),
    entry_price: z.number().positive('Entry price must be greater than 0'),
    allocation_value: z.number().positive('Allocation value must be greater than 0'),
    strategy_contrib: z.record(z.string(), z.number()),
  })).min(1, 'At least one position is required'),
});

export const updatePortfolioSchema = z.object({
  name: z.string().min(1, 'Portfolio name is required').max(255, 'Portfolio name must be at most 255 characters').optional(),
  total_investment: z.number().positive('Total investment must be greater than 0').optional(),
});

// Custom validation functions
export const validateAllocationConstraints = (constraints: AllocationConstraints): string | null => {
  if (constraints.max_allocation_per_stock <= 0 || constraints.max_allocation_per_stock > 100) {
    return 'Max allocation per stock must be between 0 and 100 percent';
  }
  
  if (constraints.min_allocation_amount < 0) {
    return 'Min allocation amount must be non-negative';
  }
  
  return null;
};

// Type inference from schemas
export type AllocationRequestFormData = z.infer<typeof allocationRequestSchema>;
export type CreatePortfolioFormData = z.infer<typeof createPortfolioSchema>;
export type UpdatePortfolioFormData = z.infer<typeof updatePortfolioSchema>;
export type AllocationConstraintsFormData = z.infer<typeof allocationConstraintsSchema>;