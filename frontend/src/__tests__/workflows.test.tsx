import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render, createMockUser, createMockStrategy, createMockStock, mockApiResponse } from '../test/utils'
import App from '../App'

// Mock all the services
const mockLogin = vi.fn()
const mockGetProfile = vi.fn()
const mockGetStrategies = vi.fn()
const mockCreateStrategy = vi.fn()
const mockGetStocks = vi.fn()
const mockGenerateAllocationPreview = vi.fn()
const mockCreatePortfolio = vi.fn()
const mockGetPortfolios = vi.fn()
const mockGetPortfolio = vi.fn()

vi.mock('../lib/api-client', () => ({
  login: (...args: any[]) => mockLogin(...args),
  getProfile: (...args: any[]) => mockGetProfile(...args),
}))

vi.mock('../services/strategyService', () => ({
  getStrategies: (...args: any[]) => mockGetStrategies(...args),
  createStrategy: (...args: any[]) => mockCreateStrategy(...args),
}))

vi.mock('../services/stockService', () => ({
  getStocks: (...args: any[]) => mockGetStocks(...args),
}))

vi.mock('../services/portfolioService', () => ({
  generateAllocationPreview: (...args: any[]) => mockGenerateAllocationPreview(...args),
  createPortfolio: (...args: any[]) => mockCreatePortfolio(...args),
  getPortfolios: (...args: any[]) => mockGetPortfolios(...args),
  getPortfolio: (...args: any[]) => mockGetPortfolio(...args),
}))

// Mock localStorage
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
}
Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage,
})

// Mock react-router-dom to control navigation
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('End-to-End Workflows', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.getItem.mockReturnValue(null)
  })

  describe('User Authentication Flow', () => {
    it('allows user to login and access dashboard', async () => {
      const mockUser = createMockUser({ name: 'John Doe', email: 'john@example.com' })
      const mockToken = 'jwt-token'

      mockLogin.mockResolvedValue(mockApiResponse({ user: mockUser, token: mockToken }))
      mockGetPortfolios.mockResolvedValue(mockApiResponse([]))

      render(<App />)

      // Should show login form initially
      expect(screen.getByText(/sign in/i)).toBeInTheDocument()

      // Fill login form
      await user.type(screen.getByLabelText(/email/i), 'john@example.com')
      await user.type(screen.getByLabelText(/password/i), 'password123')
      await user.click(screen.getByRole('button', { name: /sign in/i }))

      // Should navigate to dashboard after successful login
      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('john@example.com', 'password123')
        expect(mockLocalStorage.setItem).toHaveBeenCalledWith('token', mockToken)
      })

      // Should show navigation and user info
      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument()
        expect(screen.getByText('Dashboard')).toBeInTheDocument()
      })
    })

    it('shows error message for invalid login credentials', async () => {
      mockLogin.mockRejectedValue(new Error('Invalid credentials'))

      render(<App />)

      await user.type(screen.getByLabelText(/email/i), 'wrong@example.com')
      await user.type(screen.getByLabelText(/password/i), 'wrongpassword')
      await user.click(screen.getByRole('button', { name: /sign in/i }))

      await waitFor(() => {
        expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
      })

      expect(mockLocalStorage.setItem).not.toHaveBeenCalled()
    })
  })

  describe('Strategy Creation Workflow', () => {
    beforeEach(async () => {
      // Setup authenticated user
      const mockUser = createMockUser()
      mockLocalStorage.getItem.mockReturnValue('valid-token')
      mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))
      mockGetPortfolios.mockResolvedValue(mockApiResponse([]))
    })

    it('allows user to create a new strategy', async () => {
      const mockStrategies: any[] = []
      const mockStocks = [
        createMockStock({ id: '1', ticker: 'AAPL', name: 'Apple Inc.' }),
        createMockStock({ id: '2', ticker: 'GOOGL', name: 'Alphabet Inc.' }),
      ]
      const newStrategy = createMockStrategy({
        id: 'new-strategy-id',
        name: 'Growth Strategy',
        weightMode: 'percent',
        weightValue: 60,
      })

      mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
      mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))
      mockCreateStrategy.mockResolvedValue(mockApiResponse(newStrategy))

      render(<App />)

      // Navigate to strategies page
      await waitFor(() => {
        expect(screen.getByText('Strategies')).toBeInTheDocument()
      })
      await user.click(screen.getByText('Strategies'))

      // Wait for strategies page to load
      await waitFor(() => {
        expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
      })

      // Create new strategy
      await user.click(screen.getByText(/create new strategy/i))

      // Fill strategy form
      await user.type(screen.getByLabelText(/strategy name/i), 'Growth Strategy')
      await user.click(screen.getByLabelText(/percentage/i))
      await user.type(screen.getByLabelText(/weight value/i), '60')

      // Submit form
      await user.click(screen.getByText(/create strategy/i))

      await waitFor(() => {
        expect(mockCreateStrategy).toHaveBeenCalledWith({
          name: 'Growth Strategy',
          weightMode: 'percent',
          weightValue: 60,
        })
      })

      // Should show success message and new strategy
      await waitFor(() => {
        expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
        expect(screen.getByText('60%')).toBeInTheDocument()
      })
    })
  })

  describe('Portfolio Creation Workflow', () => {
    beforeEach(async () => {
      // Setup authenticated user
      const mockUser = createMockUser()
      mockLocalStorage.getItem.mockReturnValue('valid-token')
      mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))
    })

    it('allows user to create portfolio from strategies', async () => {
      const mockStrategies = [
        createMockStrategy({ id: '1', name: 'Growth Strategy', weightMode: 'percent', weightValue: 60 }),
        createMockStrategy({ id: '2', name: 'Value Strategy', weightMode: 'percent', weightValue: 40 }),
      ]

      const mockPreview = {
        totalInvestment: 10000,
        allocations: [
          {
            stockId: '1',
            ticker: 'AAPL',
            allocationValue: 6000,
            quantity: 40,
            price: 150.00,
            actualValue: 6000,
          },
          {
            stockId: '2',
            ticker: 'GOOGL',
            allocationValue: 4000,
            quantity: 1,
            price: 2800.00,
            actualValue: 2800,
          },
        ],
        unallocatedCash: 1200,
        totalAllocated: 8800,
      }

      const createdPortfolio = {
        id: 'new-portfolio-id',
        name: 'My Portfolio',
        totalInvestment: 10000,
      }

      mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
      mockGenerateAllocationPreview.mockResolvedValue(mockApiResponse(mockPreview))
      mockCreatePortfolio.mockResolvedValue(mockApiResponse(createdPortfolio))
      mockGetPortfolios.mockResolvedValue(mockApiResponse([]))

      render(<App />)

      // Navigate to portfolios page
      await waitFor(() => {
        expect(screen.getByText('Portfolios')).toBeInTheDocument()
      })
      await user.click(screen.getByText('Portfolios'))

      // Click create new portfolio
      await waitFor(() => {
        expect(screen.getByText(/create portfolio/i)).toBeInTheDocument()
      })
      await user.click(screen.getByText(/create portfolio/i))

      // Fill portfolio form
      await waitFor(() => {
        expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
      })

      await user.type(screen.getByLabelText(/portfolio name/i), 'My Portfolio')
      await user.type(screen.getByLabelText(/total investment/i), '10000')

      // Select strategies
      await user.click(screen.getByRole('checkbox', { name: /growth strategy/i }))
      await user.click(screen.getByRole('checkbox', { name: /value strategy/i }))

      // Generate preview
      await user.click(screen.getByText(/generate preview/i))

      // Wait for preview to load
      await waitFor(() => {
        expect(mockGenerateAllocationPreview).toHaveBeenCalledWith({
          strategyIds: ['1', '2'],
          totalInvestment: 10000,
          constraints: expect.any(Object),
        })
      })

      await waitFor(() => {
        expect(screen.getByText('AAPL')).toBeInTheDocument()
        expect(screen.getByText('GOOGL')).toBeInTheDocument()
        expect(screen.getByText('40 shares')).toBeInTheDocument()
        expect(screen.getByText('1 shares')).toBeInTheDocument()
      })

      // Create portfolio
      await user.click(screen.getByText(/create portfolio/i))

      await waitFor(() => {
        expect(mockCreatePortfolio).toHaveBeenCalledWith({
          name: 'My Portfolio',
          totalInvestment: 10000,
          positions: expect.arrayContaining([
            expect.objectContaining({
              stockId: '1',
              quantity: 40,
              entryPrice: 150.00,
            }),
            expect.objectContaining({
              stockId: '2',
              quantity: 1,
              entryPrice: 2800.00,
            }),
          ]),
        })
      })

      // Should navigate to new portfolio
      expect(mockNavigate).toHaveBeenCalledWith('/portfolios/new-portfolio-id')
    })

    it('handles allocation constraint violations gracefully', async () => {
      const mockStrategies = [createMockStrategy()]

      mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
      mockGenerateAllocationPreview.mockRejectedValue(
        new Error('Maximum allocation per stock exceeded')
      )
      mockGetPortfolios.mockResolvedValue(mockApiResponse([]))

      render(<App />)

      // Navigate to portfolio creation
      await waitFor(() => {
        expect(screen.getByText('Portfolios')).toBeInTheDocument()
      })
      await user.click(screen.getByText('Portfolios'))

      await waitFor(() => {
        expect(screen.getByText(/create portfolio/i)).toBeInTheDocument()
      })
      await user.click(screen.getByText(/create portfolio/i))

      // Fill form with conflicting constraints
      await waitFor(() => {
        expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
      })

      await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
      await user.type(screen.getByLabelText(/total investment/i), '1000')
      await user.click(screen.getByRole('checkbox', { name: /test strategy/i }))

      // Set very restrictive constraints
      await user.click(screen.getByText(/advanced settings/i))
      const maxAllocationInput = screen.getByLabelText(/maximum allocation per stock/i)
      await user.clear(maxAllocationInput)
      await user.type(maxAllocationInput, '5') // 5% max per stock

      // Try to generate preview
      await user.click(screen.getByText(/generate preview/i))

      // Should show error message
      await waitFor(() => {
        expect(screen.getByText(/maximum allocation per stock exceeded/i)).toBeInTheDocument()
      })

      // Should not proceed to portfolio creation
      expect(screen.queryByText(/create portfolio/i)).not.toBeInTheDocument()
    })
  })

  describe('Portfolio Dashboard Workflow', () => {
    it('displays portfolio performance and allows navigation', async () => {
      const mockUser = createMockUser()
      const mockPortfolio = {
        id: 'portfolio-1',
        name: 'My Investment Portfolio',
        totalInvestment: 50000,
        positions: [
          {
            portfolioId: 'portfolio-1',
            stockId: '1',
            quantity: 100,
            entryPrice: 150.00,
            allocationValue: 15000,
            currentPrice: 155.25,
            currentValue: 15525,
            pnl: 525,
            pnlPercent: 3.5,
            stock: {
              id: '1',
              ticker: 'AAPL',
              name: 'Apple Inc.',
              sector: 'Technology',
            },
          },
          {
            portfolioId: 'portfolio-1',
            stockId: '2',
            quantity: 10,
            entryPrice: 2800.00,
            allocationValue: 28000,
            currentPrice: 2750.50,
            currentValue: 27505,
            pnl: -495,
            pnlPercent: -1.77,
            stock: {
              id: '2',
              ticker: 'GOOGL',
              name: 'Alphabet Inc.',
              sector: 'Technology',
            },
          },
        ],
      }

      mockLocalStorage.getItem.mockReturnValue('valid-token')
      mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))
      mockGetPortfolios.mockResolvedValue(mockApiResponse([mockPortfolio]))
      mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))

      render(<App />)

      // Navigate to portfolios
      await waitFor(() => {
        expect(screen.getByText('Portfolios')).toBeInTheDocument()
      })
      await user.click(screen.getByText('Portfolios'))

      // Should show portfolio list
      await waitFor(() => {
        expect(screen.getByText('My Investment Portfolio')).toBeInTheDocument()
      })

      // Click on portfolio to view details
      await user.click(screen.getByText('My Investment Portfolio'))

      // Should show portfolio dashboard
      await waitFor(() => {
        expect(screen.getByText('$50,000.00')).toBeInTheDocument() // Total investment
        expect(screen.getByText('$43,030.00')).toBeInTheDocument() // Current value (15525 + 27505)
        expect(screen.getByText('$30.00')).toBeInTheDocument() // Total P&L (525 - 495)
      })

      // Should show individual positions
      expect(screen.getByText('AAPL')).toBeInTheDocument()
      expect(screen.getByText('Apple Inc.')).toBeInTheDocument()
      expect(screen.getByText('+3.50%')).toBeInTheDocument() // Positive P&L

      expect(screen.getByText('GOOGL')).toBeInTheDocument()
      expect(screen.getByText('Alphabet Inc.')).toBeInTheDocument()
      expect(screen.getByText('-1.77%')).toBeInTheDocument() // Negative P&L
    })
  })

  describe('Error Handling Workflows', () => {
    it('handles network errors gracefully', async () => {
      mockGetProfile.mockRejectedValue(new Error('Network Error'))
      mockLocalStorage.getItem.mockReturnValue('valid-token')

      render(<App />)

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })

      // Should provide retry option
      expect(screen.getByText(/try again/i)).toBeInTheDocument()
    })

    it('handles session expiration', async () => {
      mockGetProfile.mockRejectedValue(new Error('Unauthorized'))
      mockLocalStorage.getItem.mockReturnValue('expired-token')

      render(<App />)

      await waitFor(() => {
        expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('token')
      })

      // Should redirect to login
      expect(screen.getByText(/sign in/i)).toBeInTheDocument()
    })
  })
})