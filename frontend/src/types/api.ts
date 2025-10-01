import { z } from 'zod';

// Generic API response wrapper
export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

// Error response types
export interface ValidationError {
  field: string;
  tag: string;
  value: string;
  message: string;
}

export interface ApiError {
  type: 'VALIDATION_ERROR' | 'NOT_FOUND' | 'CONFLICT' | 'EXTERNAL_API_ERROR' | 'INTERNAL_ERROR';
  message: string;
  details?: Record<string, any>;
  errors?: ValidationError[];
}

// Pagination types
export interface PaginationParams {
  page: number;
  limit: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

// Search and filter types
export interface SearchParams {
  query?: string;
  filters?: Record<string, any>;
}

// Market data types
export interface Quote {
  symbol: string;
  price: number;
  change: number;
  change_percent: number;
  volume?: number;
  market_cap?: number;
  timestamp: string;
}

export interface OHLCV {
  timestamp: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface MarketDataRequest {
  symbols: string[];
  from?: string;
  to?: string;
  interval?: '1m' | '5m' | '15m' | '1h' | '1d' | '1w' | '1M';
}

// TradingView DataFeed types
export interface TradingViewSymbol {
  symbol: string;
  full_name: string;
  description: string;
  exchange: string;
  type: string;
}

export interface TradingViewSearchResult {
  symbol: string;
  full_name: string;
  description: string;
  exchange: string;
  ticker: string;
  type: string;
}

export interface TradingViewHistoryRequest {
  symbol: string;
  resolution: string;
  from: number;
  to: number;
  countback?: number;
}

export interface TradingViewHistoryResponse {
  s: 'ok' | 'no_data' | 'error';
  t?: number[];
  o?: number[];
  h?: number[];
  l?: number[];
  c?: number[];
  v?: number[];
  errmsg?: string;
}

// Authentication types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    name: string;
    email: string;
    created_at: string;
    updated_at: string;
  };
  expires_at: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

// Zod validation schemas for API types
export const paginationParamsSchema = z.object({
  page: z.number().int().positive().default(1),
  limit: z.number().int().positive().max(100).default(20),
  sort_by: z.string().optional(),
  sort_order: z.enum(['asc', 'desc']).default('asc'),
});

export const searchParamsSchema = z.object({
  query: z.string().optional(),
  filters: z.record(z.any()).optional(),
});

export const marketDataRequestSchema = z.object({
  symbols: z.array(z.string()).min(1, 'At least one symbol is required'),
  from: z.string().datetime().optional(),
  to: z.string().datetime().optional(),
  interval: z.enum(['1m', '5m', '15m', '1h', '1d', '1w', '1M']).default('1d'),
});

export const loginRequestSchema = z.object({
  email: z.string().email('Must be a valid email address'),
  password: z.string().min(1, 'Password is required'),
});

export const registerRequestSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be at most 255 characters'),
  email: z.string().email('Must be a valid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters').max(128, 'Password must be at most 128 characters'),
});

// Type inference from schemas
export type PaginationParamsFormData = z.infer<typeof paginationParamsSchema>;
export type SearchParamsFormData = z.infer<typeof searchParamsSchema>;
export type MarketDataRequestFormData = z.infer<typeof marketDataRequestSchema>;
export type LoginRequestFormData = z.infer<typeof loginRequestSchema>;
export type RegisterRequestFormData = z.infer<typeof registerRequestSchema>;

// HTTP status code helpers
export const isSuccessStatus = (status: number): boolean => status >= 200 && status < 300;
export const isClientError = (status: number): boolean => status >= 400 && status < 500;
export const isServerError = (status: number): boolean => status >= 500;

// API endpoint constants
export const API_ENDPOINTS = {
  // Authentication
  LOGIN: '/auth/login',
  REGISTER: '/auth/register',
  LOGOUT: '/auth/logout',
  REFRESH: '/auth/refresh',
  
  // Users
  USERS: '/users',
  USER_PROFILE: '/users/profile',
  
  // Strategies
  STRATEGIES: '/strategies',
  STRATEGY_STOCKS: (id: string) => `/strategies/${id}/stocks`,
  STRATEGY_WEIGHT: (id: string) => `/strategies/${id}/weight`,
  
  // Stocks
  STOCKS: '/stocks',
  STOCK_SIGNAL: (id: string) => `/stocks/${id}/signal`,
  
  // Portfolios
  PORTFOLIOS: '/portfolios',
  PORTFOLIO_PREVIEW: '/portfolios/preview',
  PORTFOLIO_HISTORY: (id: string) => `/portfolios/${id}/history`,
  
  // Market Data
  QUOTES: '/quotes',
  QUOTE: (symbol: string) => `/quotes/${symbol}`,
  OHLCV: (symbol: string) => `/ohlcv/${symbol}`,
  
  // TradingView DataFeed
  TRADINGVIEW_SYMBOLS: '/tradingview/symbols',
  TRADINGVIEW_SEARCH: '/tradingview/search',
  TRADINGVIEW_HISTORY: '/tradingview/history',
} as const;