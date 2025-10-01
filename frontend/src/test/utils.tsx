import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from '../contexts/AuthContext'

// Create a custom render function that includes providers
const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          {children}
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
) => render(ui, { wrapper: AllTheProviders, ...options })

export * from '@testing-library/react'
export { customRender as render }

// Mock data factories
export const createMockUser = (overrides = {}) => ({
  id: '1',
  name: 'Test User',
  email: 'test@example.com',
  ...overrides,
})

export const createMockStrategy = (overrides = {}) => ({
  id: '1',
  name: 'Test Strategy',
  weightMode: 'percent' as const,
  weightValue: 50,
  userId: '1',
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

export const createMockStock = (overrides = {}) => ({
  id: '1',
  ticker: 'AAPL',
  name: 'Apple Inc.',
  sector: 'Technology',
  createdAt: new Date().toISOString(),
  ...overrides,
})

export const createMockPortfolio = (overrides = {}) => ({
  id: '1',
  name: 'Test Portfolio',
  totalInvestment: 10000,
  userId: '1',
  positions: [],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  ...overrides,
})

export const createMockPosition = (overrides = {}) => ({
  portfolioId: '1',
  stockId: '1',
  quantity: 100,
  entryPrice: 150.50,
  allocationValue: 15050,
  stock: createMockStock(),
  currentPrice: 155.25,
  currentValue: 15525,
  pnl: 475,
  pnlPercent: 3.15,
  ...overrides,
})

export const createMockQuote = (overrides = {}) => ({
  symbol: 'AAPL',
  price: 150.50,
  change: 2.25,
  changePercent: 1.52,
  volume: 1000000,
  high: 152.00,
  low: 148.75,
  open: 149.25,
  previousClose: 148.25,
  timestamp: new Date().toISOString(),
  ...overrides,
})

// Mock API responses
export function mockApiResponse<T>(data: T, delay = 0): Promise<T> {
  return new Promise<T>((resolve) => {
    setTimeout(() => resolve(data), delay)
  })
}

export const mockApiError = (message = 'API Error', status = 500, delay = 0) => {
  return new Promise((_, reject) => {
    setTimeout(() => {
      const error = new Error(message) as any
      error.response = { status, data: { message } }
      reject(error)
    }, delay)
  })
}