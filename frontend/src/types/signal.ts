import { z } from 'zod';
import type { Stock } from './stock';

// Signal types
export type SignalType = 'Buy' | 'Hold';

export interface Signal {
  stock_id: string;
  signal: SignalType;
  date: string;
  created_at: string;
  stock?: Stock;
}

export interface CreateSignalRequest {
  stock_id: string;
  signal: SignalType;
  date: string;
}

export interface UpdateSignalRequest {
  signal: SignalType;
}

export interface SignalResponse {
  stock_id: string;
  signal: SignalType;
  date: string;
  created_at: string;
  stock?: Stock;
}

// Zod validation schemas
export const signalTypeSchema = z.enum(['Buy', 'Hold'], {
  errorMap: () => ({ message: 'Signal must be either Buy or Hold' }),
});

export const createSignalSchema = z.object({
  stock_id: z.string().uuid('Stock ID must be a valid UUID'),
  signal: signalTypeSchema,
  date: z.string().datetime('Date must be a valid ISO datetime string'),
});

export const updateSignalSchema = z.object({
  signal: signalTypeSchema,
});

// Type inference from schemas
export type CreateSignalFormData = z.infer<typeof createSignalSchema>;
export type UpdateSignalFormData = z.infer<typeof updateSignalSchema>;