import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { render, createMockPortfolio, createMockPosition, mockApiResponse } from '../../test/utils'
import PortfolioDashboard from '../PortfolioDashboard'

// Mock the portfolio service
const mockGetPortfolio = vi.fn()
const mockGetPortfolioHistory = vi.fn()
vi.mock('../../services/portfolioService', () => ({
  getPortfolio: (...args: any[]) => mockGetPortfolio(...args),
  getPortfolioHistory: (...args: any[]) => mockGetPortfolioHistory(...args),
}))

// Mock recharts to avoid canvas issues in tests
vi.mock('recharts', () => ({
  LineChart: ({ children }: any) => <div data-testid="line-chart">{children}</div>,
  Line: () => <div data-testid="line" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  ResponsiveContainer: ({ children }: any) => <div data-testid="responsive-container">{children}</div>,
}))

describe('PortfolioDashboard', () => {
  const mockPortfolioId = 'portfolio-1'

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders portfolio information when data is loaded', async () => {
    const mockPortfolio = createMockPortfolio({
      id: mockPortfolioId,
      name: 'My Test Portfolio',
      totalInvestment: 50000,
      positions: [
        createMockPosition({
          stock: { ticker: 'AAPL', name: 'Apple Inc.' },
          quantity: 100,
          currentValue: 15500,
          pnl: 500,
          pnlPercent: 3.33,
        }),
        createMockPosition({
          stock: { ticker: 'GOOGL', name: 'Alphabet Inc.' },
          quantity: 10,
          currentValue: 28000,
          pnl: -200,
          pnlPercent: -0.71,
        }),
      ],
    })

    const mockHistory = [
      {
        timestamp: '2024-01-01T00:00:00Z',
        nav: 50000,
        pnl: 0,
        drawdown: 0,
      },
      {
        timestamp: '2024-01-02T00:00:00Z',
        nav: 50300,
        pnl: 300,
        drawdown: 0,
      },
    ]

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse(mockHistory))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('My Test Portfolio')).toBeInTheDocument()
    })

    // Check portfolio summary
    expect(screen.getByText('$50,000.00')).toBeInTheDocument() // Total investment
    expect(screen.getByText('$43,500.00')).toBeInTheDocument() // Current value (15500 + 28000)
    expect(screen.getByText('$300.00')).toBeInTheDocument() // Total P&L (500 - 200)

    // Check positions
    expect(screen.getByText('AAPL')).toBeInTheDocument()
    expect(screen.getByText('Apple Inc.')).toBeInTheDocument()
    expect(screen.getByText('GOOGL')).toBeInTheDocument()
    expect(screen.getByText('Alphabet Inc.')).toBeInTheDocument()

    // Check chart is rendered
    expect(screen.getByTestId('line-chart')).toBeInTheDocument()
  })

  it('shows loading state while fetching data', () => {
    mockGetPortfolio.mockReturnValue(new Promise(() => {})) // Never resolves
    mockGetPortfolioHistory.mockReturnValue(new Promise(() => {}))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
  })

  it('shows error state when portfolio fetch fails', async () => {
    const errorMessage = 'Failed to fetch portfolio'
    mockGetPortfolio.mockRejectedValue(new Error(errorMessage))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse([]))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument()
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
  })

  it('handles empty portfolio positions', async () => {
    const mockPortfolio = createMockPortfolio({
      id: mockPortfolioId,
      name: 'Empty Portfolio',
      positions: [],
    })

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse([]))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText('Empty Portfolio')).toBeInTheDocument()
    })

    expect(screen.getByText(/no positions/i)).toBeInTheDocument()
  })

  it('calculates and displays performance metrics correctly', async () => {
    const mockPortfolio = createMockPortfolio({
      totalInvestment: 10000,
      positions: [
        createMockPosition({
          currentValue: 11000,
          pnl: 1000,
          pnlPercent: 10.0,
        }),
      ],
    })

    const mockHistory = [
      {
        timestamp: '2024-01-01T00:00:00Z',
        nav: 10000,
        pnl: 0,
        drawdown: 0,
      },
      {
        timestamp: '2024-01-02T00:00:00Z',
        nav: 12000,
        pnl: 2000,
        drawdown: 0,
      },
      {
        timestamp: '2024-01-03T00:00:00Z',
        nav: 11000,
        pnl: 1000,
        drawdown: -8.33, // (11000 - 12000) / 12000 * 100
      },
    ]

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse(mockHistory))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText('10.00%')).toBeInTheDocument() // Total return percentage
    })

    // Should show max drawdown
    expect(screen.getByText('-8.33%')).toBeInTheDocument()
  })

  it('formats currency values correctly', async () => {
    const mockPortfolio = createMockPortfolio({
      totalInvestment: 1234567.89,
      positions: [
        createMockPosition({
          currentValue: 1300000,
          pnl: 65432.11,
        }),
      ],
    })

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse([]))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText('$1,234,567.89')).toBeInTheDocument()
      expect(screen.getByText('$65,432.11')).toBeInTheDocument()
    })
  })

  it('shows positive and negative P&L with correct styling', async () => {
    const mockPortfolio = createMockPortfolio({
      positions: [
        createMockPosition({
          stock: { ticker: 'WINNER' },
          pnl: 1000,
          pnlPercent: 10.0,
        }),
        createMockPosition({
          stock: { ticker: 'LOSER' },
          pnl: -500,
          pnlPercent: -5.0,
        }),
      ],
    })

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse([]))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText('WINNER')).toBeInTheDocument()
      expect(screen.getByText('LOSER')).toBeInTheDocument()
    })

    // Check for positive P&L styling (green)
    const positiveElement = screen.getByText('+10.00%')
    expect(positiveElement).toHaveClass('text-green-600')

    // Check for negative P&L styling (red)
    const negativeElement = screen.getByText('-5.00%')
    expect(negativeElement).toHaveClass('text-red-600')
  })

  it('refreshes data when portfolio ID changes', async () => {
    const { rerender } = render(<PortfolioDashboard portfolioId="portfolio-1" />)

    expect(mockGetPortfolio).toHaveBeenCalledWith('portfolio-1')

    // Change portfolio ID
    rerender(<PortfolioDashboard portfolioId="portfolio-2" />)

    expect(mockGetPortfolio).toHaveBeenCalledWith('portfolio-2')
    expect(mockGetPortfolio).toHaveBeenCalledTimes(2)
  })

  it('handles stale market data indicators', async () => {
    const mockPortfolio = createMockPortfolio({
      positions: [
        createMockPosition({
          stock: { ticker: 'AAPL' },
          // Simulate stale data (older than 15 minutes)
          lastUpdated: new Date(Date.now() - 20 * 60 * 1000).toISOString(),
        }),
      ],
    })

    mockGetPortfolio.mockResolvedValue(mockApiResponse(mockPortfolio))
    mockGetPortfolioHistory.mockResolvedValue(mockApiResponse([]))

    render(<PortfolioDashboard portfolioId={mockPortfolioId} />)

    await waitFor(() => {
      expect(screen.getByText(/stale data/i)).toBeInTheDocument()
    })
  })
})