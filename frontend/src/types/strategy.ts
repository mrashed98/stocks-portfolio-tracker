import { z } from 'zod';
import type { StrategyStock } from './stock';

// Strategy types
export type WeightMode = 'percent' | 'budget';

export interface Strategy {
  id: string;
  user_id: string;
  name: string;
  weight_mode: WeightMode;
  weight_value: number;
  created_at: string;
  updated_at: string;
  stocks?: StrategyStock[];
}

export interface CreateStrategyRequest {
  name: string;
  weight_mode: WeightMode;
  weight_value: number;
}

export interface UpdateStrategyRequest {
  name?: string;
  weight_mode?: WeightMode;
  weight_value?: number;
}

export interface StrategyResponse {
  id: string;
  user_id: string;
  name: string;
  weight_mode: WeightMode;
  weight_value: number;
  created_at: string;
  updated_at: string;
  stocks?: StrategyStock[];
}

// Zod validation schemas
export const weightModeSchema = z.enum(['percent', 'budget'], {
  errorMap: () => ({ message: 'Weight mode must be either percent or budget' }),
});

export const createStrategySchema = z.object({
  name: z.string().min(1, 'Strategy name is required').max(255, 'Strategy name must be at most 255 characters'),
  weight_mode: weightModeSchema,
  weight_value: z.number().positive('Weight value must be greater than 0'),
});

export const updateStrategySchema = z.object({
  name: z.string().min(1, 'Strategy name is required').max(255, 'Strategy name must be at most 255 characters').optional(),
  weight_mode: weightModeSchema.optional(),
  weight_value: z.number().positive('Weight value must be greater than 0').optional(),
});

// Custom validation for strategy weights
export const validateStrategyWeights = (strategies: Strategy[]): string | null => {
  const percentageStrategies = strategies.filter(s => s.weight_mode === 'percent');
  const totalPercentage = percentageStrategies.reduce((sum, strategy) => sum + strategy.weight_value, 0);
  
  if (totalPercentage > 100) {
    return `Total percentage weights cannot exceed 100%, current total: ${totalPercentage.toFixed(2)}%`;
  }
  
  return null;
};

// Type inference from schemas
export type CreateStrategyFormData = z.infer<typeof createStrategySchema>;
export type UpdateStrategyFormData = z.infer<typeof updateStrategySchema>;