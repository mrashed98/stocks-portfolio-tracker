export const config = {
  apiUrl: (import.meta.env.VITE_API_URL as string) || 'http://localhost:8080',
  isDevelopment: import.meta.env.DEV as boolean,
  isProduction: import.meta.env.PROD as boolean,
} as const

export const apiEndpoints = {
  health: '/health',
  auth: {
    login: '/api/v1/auth/login',
    register: '/api/v1/auth/register',
    refresh: '/api/v1/auth/refresh',
  },
  strategies: {
    list: '/api/v1/strategies',
    create: '/api/v1/strategies',
    update: (id: string) => `/api/v1/strategies/${id}`,
    delete: (id: string) => `/api/v1/strategies/${id}`,
    stocks: (id: string) => `/api/v1/strategies/${id}/stocks`,
    weight: (id: string) => `/api/v1/strategies/${id}/weight`,
  },
  stocks: {
    list: '/api/v1/stocks',
    create: '/api/v1/stocks',
    update: (id: string) => `/api/v1/stocks/${id}`,
    signal: (id: string) => `/api/v1/stocks/${id}/signal`,
  },
  portfolios: {
    list: '/api/v1/portfolios',
    create: '/api/v1/portfolios',
    preview: '/api/v1/portfolios/preview',
    detail: (id: string) => `/api/v1/portfolios/${id}`,
    history: (id: string) => `/api/v1/portfolios/${id}/history`,
  },
  quotes: {
    single: (ticker: string) => `/api/v1/quotes/${ticker}`,
    batch: '/api/v1/quotes/batch',
    ohlcv: (ticker: string) => `/api/v1/ohlcv/${ticker}`,
  },
  tradingview: {
    symbols: '/api/v1/tradingview/symbols',
    history: '/api/v1/tradingview/history',
    search: '/api/v1/tradingview/search',
  },
} as const