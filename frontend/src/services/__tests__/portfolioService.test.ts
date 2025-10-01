import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as portfolioService from '../portfolioService'
import { mockApiResponse, mockApiError, createMockPortfolio, createMockPosition } from '../../test/utils'

// Mock the API client
const mockApiClient = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}

vi.mock('../../lib/api-client', () => ({
  default: mockApiClient,
}))

describe('portfolioService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getPortfolios', () => {
    it('fetches user portfolios successfully', async () => {
      const mockPortfolios = [
        createMockPortfolio({ id: '1', name: 'Portfolio 1' }),
        createMockPortfolio({ id: '2', name: 'Portfolio 2' }),
      ]

      mockApiClient.get.mockResolvedValue({ data: mockPortfolios })

      const result = await portfolioService.getPortfolios()

      expect(mockApiClient.get).toHaveBeenCalledWith('/portfolios')
      expect(result).toEqual(mockPortfolios)
    })

    it('handles API errors', async () => {
      const errorMessage = 'Failed to fetch portfolios'
      mockApiClient.get.mockRejectedValue(mockApiError(errorMessage))

      await expect(portfolioService.getPortfolios()).rejects.toThrow(errorMessage)
    })
  })

  describe('getPortfolio', () => {
    it('fetches single portfolio successfully', async () => {
      const mockPortfolio = createMockPortfolio({
        id: '1',
        name: 'Test Portfolio',
        positions: [
          createMockPosition({ stock: { ticker: 'AAPL' } }),
          createMockPosition({ stock: { ticker: 'GOOGL' } }),
        ],
      })

      mockApiClient.get.mockResolvedValue({ data: mockPortfolio })

      const result = await portfolioService.getPortfolio('1')

      expect(mockApiClient.get).toHaveBeenCalledWith('/portfolios/1')
      expect(result).toEqual(mockPortfolio)
      expect(result.positions).toHaveLength(2)
    })

    it('handles portfolio not found', async () => {
      mockApiClient.get.mockRejectedValue(mockApiError('Portfolio not found', 404))

      await expect(portfolioService.getPortfolio('nonexistent')).rejects.toThrow('Portfolio not found')
    })
  })

  describe('createPortfolio', () => {
    it('creates portfolio successfully', async () => {
      const portfolioData = {
        name: 'New Portfolio',
        totalInvestment: 10000,
        positions: [
          {
            stockId: '1',
            quantity: 100,
            entryPrice: 150.50,
            allocationValue: 15050,
            strategyContrib: { 'strategy-1': 15050 },
          },
        ],
      }

      const createdPortfolio = createMockPortfolio({
        id: 'new-portfolio-id',
        ...portfolioData,
      })

      mockApiClient.post.mockResolvedValue({ data: createdPortfolio })

      const result = await portfolioService.createPortfolio(portfolioData)

      expect(mockApiClient.post).toHaveBeenCalledWith('/portfolios', portfolioData)
      expect(result).toEqual(createdPortfolio)
    })

    it('validates portfolio data before creation', async () => {
      const invalidData = {
        name: '', // Empty name
        totalInvestment: -1000, // Negative investment
        positions: [], // No positions
      }

      await expect(portfolioService.createPortfolio(invalidData as any)).rejects.toThrow(/validation/i)
      expect(mockApiClient.post).not.toHaveBeenCalled()
    })
  })

  describe('generateAllocationPreview', () => {
    it('generates allocation preview successfully', async () => {
      const request = {
        strategyIds: ['strategy-1', 'strategy-2'],
        totalInvestment: 10000,
        constraints: {
          maxAllocationPerStock: 25,
          minAllocationAmount: 100,
        },
      }

      const mockPreview = {
        totalInvestment: 10000,
        allocations: [
          {
            stockId: '1',
            ticker: 'AAPL',
            allocationValue: 5000,
            quantity: 33,
            price: 150.50,
            actualValue: 4966.50,
          },
          {
            stockId: '2',
            ticker: 'GOOGL',
            allocationValue: 5000,
            quantity: 1,
            price: 2800.75,
            actualValue: 2800.75,
          },
        ],
        unallocatedCash: 232.75,
        totalAllocated: 7767.25,
      }

      mockApiClient.post.mockResolvedValue({ data: mockPreview })

      const result = await portfolioService.generateAllocationPreview(request)

      expect(mockApiClient.post).toHaveBeenCalledWith('/portfolios/preview', request)
      expect(result).toEqual(mockPreview)
      expect(result.allocations).toHaveLength(2)
    })

    it('handles allocation constraint violations', async () => {
      const request = {
        strategyIds: ['strategy-1'],
        totalInvestment: 1000,
        constraints: {
          maxAllocationPerStock: 10, // Very restrictive
          minAllocationAmount: 500, // Conflicting with max
        },
      }

      mockApiClient.post.mockRejectedValue(
        mockApiError('Allocation constraints cannot be satisfied', 422)
      )

      await expect(portfolioService.generateAllocationPreview(request)).rejects.toThrow(
        'Allocation constraints cannot be satisfied'
      )
    })
  })

  describe('updatePortfolio', () => {
    it('updates portfolio successfully', async () => {
      const portfolioId = 'portfolio-1'
      const updates = {
        name: 'Updated Portfolio Name',
        totalInvestment: 15000,
      }

      const updatedPortfolio = createMockPortfolio({
        id: portfolioId,
        ...updates,
      })

      mockApiClient.put.mockResolvedValue({ data: updatedPortfolio })

      const result = await portfolioService.updatePortfolio(portfolioId, updates)

      expect(mockApiClient.put).toHaveBeenCalledWith(`/portfolios/${portfolioId}`, updates)
      expect(result).toEqual(updatedPortfolio)
    })
  })

  describe('deletePortfolio', () => {
    it('deletes portfolio successfully', async () => {
      const portfolioId = 'portfolio-1'

      mockApiClient.delete.mockResolvedValue({ data: { success: true } })

      await portfolioService.deletePortfolio(portfolioId)

      expect(mockApiClient.delete).toHaveBeenCalledWith(`/portfolios/${portfolioId}`)
    })

    it('handles deletion of non-existent portfolio', async () => {
      mockApiClient.delete.mockRejectedValue(mockApiError('Portfolio not found', 404))

      await expect(portfolioService.deletePortfolio('nonexistent')).rejects.toThrow('Portfolio not found')
    })
  })

  describe('getPortfolioHistory', () => {
    it('fetches portfolio NAV history successfully', async () => {
      const portfolioId = 'portfolio-1'
      const from = new Date('2024-01-01')
      const to = new Date('2024-01-31')

      const mockHistory = [
        {
          timestamp: '2024-01-01T00:00:00Z',
          nav: 10000,
          pnl: 0,
          drawdown: 0,
        },
        {
          timestamp: '2024-01-15T00:00:00Z',
          nav: 10500,
          pnl: 500,
          drawdown: 0,
        },
        {
          timestamp: '2024-01-31T00:00:00Z',
          nav: 10200,
          pnl: 200,
          drawdown: -2.86,
        },
      ]

      mockApiClient.get.mockResolvedValue({ data: mockHistory })

      const result = await portfolioService.getPortfolioHistory(portfolioId, from, to)

      expect(mockApiClient.get).toHaveBeenCalledWith(
        `/portfolios/${portfolioId}/history`,
        {
          params: {
            from: from.toISOString(),
            to: to.toISOString(),
          },
        }
      )
      expect(result).toEqual(mockHistory)
      expect(result).toHaveLength(3)
    })

    it('handles empty history gracefully', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] })

      const result = await portfolioService.getPortfolioHistory('portfolio-1', new Date(), new Date())

      expect(result).toEqual([])
    })
  })

  describe('rebalancePortfolio', () => {
    it('rebalances portfolio successfully', async () => {
      const portfolioId = 'portfolio-1'
      const newTotalInvestment = 20000

      const rebalancedPortfolio = createMockPortfolio({
        id: portfolioId,
        totalInvestment: newTotalInvestment,
      })

      mockApiClient.post.mockResolvedValue({ data: rebalancedPortfolio })

      const result = await portfolioService.rebalancePortfolio(portfolioId, newTotalInvestment)

      expect(mockApiClient.post).toHaveBeenCalledWith(
        `/portfolios/${portfolioId}/rebalance`,
        { newTotalInvestment }
      )
      expect(result).toEqual(rebalancedPortfolio)
    })
  })

  describe('error handling', () => {
    it('handles network errors', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network Error'))

      await expect(portfolioService.getPortfolios()).rejects.toThrow('Network Error')
    })

    it('handles timeout errors', async () => {
      mockApiClient.get.mockRejectedValue(mockApiError('Request timeout', 408))

      await expect(portfolioService.getPortfolios()).rejects.toThrow('Request timeout')
    })

    it('handles server errors', async () => {
      mockApiClient.get.mockRejectedValue(mockApiError('Internal server error', 500))

      await expect(portfolioService.getPortfolios()).rejects.toThrow('Internal server error')
    })
  })

  describe('data transformation', () => {
    it('transforms API response data correctly', async () => {
      const apiResponse = {
        id: '1',
        name: 'Test Portfolio',
        total_investment: 10000, // Snake case from API
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        positions: [
          {
            portfolio_id: '1',
            stock_id: '1',
            quantity: 100,
            entry_price: 150.50,
            allocation_value: 15050,
            current_price: 155.25,
            current_value: 15525,
            pnl: 475,
            pnl_percent: 3.15,
            stock: {
              id: '1',
              ticker: 'AAPL',
              name: 'Apple Inc.',
              sector: 'Technology',
            },
          },
        ],
      }

      mockApiClient.get.mockResolvedValue({ data: apiResponse })

      const result = await portfolioService.getPortfolio('1')

      // Should transform snake_case to camelCase
      expect(result.totalInvestment).toBe(10000)
      expect(result.createdAt).toBe('2024-01-01T00:00:00Z')
      expect(result.positions[0].entryPrice).toBe(150.50)
      expect(result.positions[0].currentPrice).toBe(155.25)
    })
  })
})