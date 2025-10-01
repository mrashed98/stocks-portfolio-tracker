import { z } from 'zod';
import type { Stock } from './stock';

// Position interfaces
export interface Position {
  portfolio_id: string;
  stock_id: string;
  quantity: number;
  entry_price: number;
  allocation_value: number;
  strategy_contrib: Record<string, number>;
  created_at: string;
  updated_at: string;
  stock?: Stock;
  current_price?: number;
  current_value?: number;
  pnl?: number;
  pnl_percentage?: number;
}

export interface CreatePositionRequest {
  stock_id: string;
  quantity: number;
  entry_price: number;
  allocation_value: number;
  strategy_contrib: Record<string, number>;
}

export interface UpdatePositionRequest {
  quantity?: number;
  entry_price?: number;
  allocation_value?: number;
}

export interface PositionResponse {
  portfolio_id: string;
  stock_id: string;
  quantity: number;
  entry_price: number;
  allocation_value: number;
  strategy_contrib: Record<string, number>;
  created_at: string;
  updated_at: string;
  stock?: Stock;
  current_price?: number;
  current_value?: number;
  pnl?: number;
  pnl_percentage?: number;
}

// Zod validation schemas
export const createPositionSchema = z.object({
  stock_id: z.string().uuid('Stock ID must be a valid UUID'),
  quantity: z.number().int().positive('Quantity must be a positive integer'),
  entry_price: z.number().positive('Entry price must be greater than 0'),
  allocation_value: z.number().positive('Allocation value must be greater than 0'),
  strategy_contrib: z.record(z.string(), z.number()),
});

export const updatePositionSchema = z.object({
  quantity: z.number().int().positive('Quantity must be a positive integer').optional(),
  entry_price: z.number().positive('Entry price must be greater than 0').optional(),
  allocation_value: z.number().positive('Allocation value must be greater than 0').optional(),
});

// Helper functions for position calculations
export const calculatePositionMetrics = (
  position: Position,
  currentPrice: number
): {
  current_value: number;
  pnl: number;
  pnl_percentage: number;
} => {
  const currentValue = currentPrice * position.quantity;
  const pnl = currentValue - position.allocation_value;
  const pnlPercentage = position.allocation_value > 0 ? (pnl / position.allocation_value) * 100 : 0;
  
  return {
    current_value: currentValue,
    pnl,
    pnl_percentage: pnlPercentage,
  };
};

// Type inference from schemas
export type CreatePositionFormData = z.infer<typeof createPositionSchema>;
export type UpdatePositionFormData = z.infer<typeof updatePositionSchema>;