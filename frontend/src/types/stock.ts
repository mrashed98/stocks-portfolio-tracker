import { z } from 'zod';
import type { Signal } from './signal';

// Stock interfaces
export interface Stock {
  id: string;
  ticker: string;
  name: string;
  sector?: string;
  exchange?: string;
  created_at: string;
  updated_at: string;
  current_signal?: Signal;
}

export interface CreateStockRequest {
  ticker: string;
  name: string;
  sector?: string;
  exchange?: string;
}

export interface UpdateStockRequest {
  name?: string;
  sector?: string;
  exchange?: string;
}

export interface StockResponse {
  id: string;
  ticker: string;
  name: string;
  sector?: string;
  exchange?: string;
  created_at: string;
  updated_at: string;
  current_signal?: Signal;
}

// Strategy-Stock relationship
export interface StrategyStock {
  strategy_id: string;
  stock_id: string;
  eligible: boolean;
  created_at: string;
  stock?: Stock;
}

export interface UpdateStockEligibilityRequest {
  eligible: boolean;
}

// Zod validation schemas
export const createStockSchema = z.object({
  ticker: z.string()
    .min(1, 'Ticker is required')
    .max(20, 'Ticker must be at most 20 characters')
    .regex(/^[A-Z]+$/, 'Ticker must contain only uppercase letters'),
  name: z.string().min(1, 'Stock name is required').max(255, 'Stock name must be at most 255 characters'),
  sector: z.string().max(100, 'Sector must be at most 100 characters').optional(),
  exchange: z.string().max(50, 'Exchange must be at most 50 characters').optional(),
});

export const updateStockSchema = z.object({
  name: z.string().min(1, 'Stock name is required').max(255, 'Stock name must be at most 255 characters').optional(),
  sector: z.string().max(100, 'Sector must be at most 100 characters').optional(),
  exchange: z.string().max(50, 'Exchange must be at most 50 characters').optional(),
});

export const updateStockEligibilitySchema = z.object({
  eligible: z.boolean(),
});

// Type inference from schemas
export type CreateStockFormData = z.infer<typeof createStockSchema>;
export type UpdateStockFormData = z.infer<typeof updateStockSchema>;
export type UpdateStockEligibilityFormData = z.infer<typeof updateStockEligibilitySchema>;