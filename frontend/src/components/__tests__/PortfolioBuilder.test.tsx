import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render, createMockStrategy, createMockStock, mockApiResponse } from '../../test/utils'
import PortfolioBuilder from '../PortfolioBuilder'

// Mock the services
const mockGetStrategies = vi.fn()
const mockGenerateAllocationPreview = vi.fn()
const mockCreatePortfolio = vi.fn()
vi.mock('../../services/strategyService', () => ({
  getStrategies: (...args: any[]) => mockGetStrategies(...args),
}))

vi.mock('../../services/portfolioService', () => ({
  generateAllocationPreview: (...args: any[]) => mockGenerateAllocationPreview(...args),
  createPortfolio: (...args: any[]) => mockCreatePortfolio(...args),
}))

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('PortfolioBuilder', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders portfolio builder form', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy' }),
      createMockStrategy({ id: '2', name: 'Value Strategy' }),
    ]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByText('Create New Portfolio')).toBeInTheDocument()
    })

    expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/total investment/i)).toBeInTheDocument()
    expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    expect(screen.getByText('Value Strategy')).toBeInTheDocument()
  })

  it('allows user to input portfolio details', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    // Fill in portfolio name
    const nameInput = screen.getByLabelText(/portfolio name/i)
    await user.type(nameInput, 'My New Portfolio')
    expect(nameInput).toHaveValue('My New Portfolio')

    // Fill in investment amount
    const investmentInput = screen.getByLabelText(/total investment/i)
    await user.type(investmentInput, '50000')
    expect(investmentInput).toHaveValue('50000')
  })

  it('allows user to select strategies', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy' }),
      createMockStrategy({ id: '2', name: 'Value Strategy' }),
    ]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    })

    // Select strategies
    const growthCheckbox = screen.getByRole('checkbox', { name: /growth strategy/i })
    const valueCheckbox = screen.getByRole('checkbox', { name: /value strategy/i })

    await user.click(growthCheckbox)
    await user.click(valueCheckbox)

    expect(growthCheckbox).toBeChecked()
    expect(valueCheckbox).toBeChecked()
  })

  it('generates allocation preview when form is valid', async () => {
    const mockStrategies = [createMockStrategy({ id: '1', name: 'Growth Strategy' })]
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

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGenerateAllocationPreview.mockResolvedValue(mockApiResponse(mockPreview))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    // Fill form
    await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
    await user.type(screen.getByLabelText(/total investment/i), '10000')
    await user.click(screen.getByRole('checkbox', { name: /growth strategy/i }))

    // Generate preview
    const previewButton = screen.getByText(/generate preview/i)
    await user.click(previewButton)

    await waitFor(() => {
      expect(mockGenerateAllocationPreview).toHaveBeenCalledWith({
        strategyIds: ['1'],
        totalInvestment: 10000,
        constraints: expect.any(Object),
      })
    })

    // Check preview results
    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeInTheDocument()
      expect(screen.getByText('GOOGL')).toBeInTheDocument()
      expect(screen.getByText('33 shares')).toBeInTheDocument()
      expect(screen.getByText('1 shares')).toBeInTheDocument()
    })
  })

  it('shows validation errors for invalid input', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByText(/generate preview/i)).toBeInTheDocument()
    })

    // Try to generate preview without filling required fields
    const previewButton = screen.getByText(/generate preview/i)
    await user.click(previewButton)

    await waitFor(() => {
      expect(screen.getByText(/portfolio name is required/i)).toBeInTheDocument()
      expect(screen.getByText(/total investment is required/i)).toBeInTheDocument()
      expect(screen.getByText(/at least one strategy must be selected/i)).toBeInTheDocument()
    })
  })

  it('validates investment amount constraints', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/total investment/i)).toBeInTheDocument()
    })

    // Test minimum investment validation
    const investmentInput = screen.getByLabelText(/total investment/i)
    await user.type(investmentInput, '50') // Below minimum

    const previewButton = screen.getByText(/generate preview/i)
    await user.click(previewButton)

    await waitFor(() => {
      expect(screen.getByText(/minimum investment is \$100/i)).toBeInTheDocument()
    })
  })

  it('allows user to modify allocation constraints', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByText(/advanced settings/i)).toBeInTheDocument()
    })

    // Open advanced settings
    const advancedButton = screen.getByText(/advanced settings/i)
    await user.click(advancedButton)

    // Modify constraints
    const maxAllocationInput = screen.getByLabelText(/maximum allocation per stock/i)
    await user.clear(maxAllocationInput)
    await user.type(maxAllocationInput, '25')

    const minAllocationInput = screen.getByLabelText(/minimum allocation amount/i)
    await user.clear(minAllocationInput)
    await user.type(minAllocationInput, '500')

    expect(maxAllocationInput).toHaveValue('25')
    expect(minAllocationInput).toHaveValue('500')
  })

  it('creates portfolio when user confirms allocation', async () => {
    const mockStrategies = [createMockStrategy({ id: '1' })]
    const mockPreview = {
      totalInvestment: 10000,
      allocations: [
        {
          stockId: '1',
          ticker: 'AAPL',
          allocationValue: 10000,
          quantity: 66,
          price: 150.50,
          actualValue: 9933.00,
        },
      ],
      unallocatedCash: 67.00,
      totalAllocated: 9933.00,
    }
    const mockCreatedPortfolio = { id: 'new-portfolio-id', name: 'Test Portfolio' }

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGenerateAllocationPreview.mockResolvedValue(mockApiResponse(mockPreview))
    mockCreatePortfolio.mockResolvedValue(mockApiResponse(mockCreatedPortfolio))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    // Fill form and generate preview
    await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
    await user.type(screen.getByLabelText(/total investment/i), '10000')
    await user.click(screen.getByRole('checkbox', { name: /test strategy/i }))
    await user.click(screen.getByText(/generate preview/i))

    // Wait for preview and create portfolio
    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeInTheDocument()
    })

    const createButton = screen.getByText(/create portfolio/i)
    await user.click(createButton)

    await waitFor(() => {
      expect(mockCreatePortfolio).toHaveBeenCalledWith({
        name: 'Test Portfolio',
        totalInvestment: 10000,
        positions: expect.arrayContaining([
          expect.objectContaining({
            stockId: '1',
            quantity: 66,
            entryPrice: 150.50,
            allocationValue: 10000,
          }),
        ]),
      })
    })

    // Should navigate to new portfolio
    expect(mockNavigate).toHaveBeenCalledWith('/portfolios/new-portfolio-id')
  })

  it('handles allocation preview errors gracefully', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGenerateAllocationPreview.mockRejectedValue(new Error('Allocation constraints violated'))

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    // Fill form and try to generate preview
    await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
    await user.type(screen.getByLabelText(/total investment/i), '10000')
    await user.click(screen.getByRole('checkbox', { name: /test strategy/i }))
    await user.click(screen.getByText(/generate preview/i))

    await waitFor(() => {
      expect(screen.getByText(/allocation constraints violated/i)).toBeInTheDocument()
    })
  })

  it('allows user to exclude stocks from allocation', async () => {
    const mockStrategies = [createMockStrategy()]
    const mockPreview = {
      totalInvestment: 10000,
      allocations: [
        {
          stockId: '1',
          ticker: 'AAPL',
          allocationValue: 5000,
          quantity: 33,
          price: 150.50,
        },
        {
          stockId: '2',
          ticker: 'GOOGL',
          allocationValue: 5000,
          quantity: 1,
          price: 2800.75,
        },
      ],
      unallocatedCash: 0,
      totalAllocated: 10000,
    }

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGenerateAllocationPreview.mockResolvedValue(mockApiResponse(mockPreview))

    render(<PortfolioBuilder />)

    // Generate initial preview
    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
    await user.type(screen.getByLabelText(/total investment/i), '10000')
    await user.click(screen.getByRole('checkbox', { name: /test strategy/i }))
    await user.click(screen.getByText(/generate preview/i))

    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeInTheDocument()
      expect(screen.getByText('GOOGL')).toBeInTheDocument()
    })

    // Exclude GOOGL
    const excludeButton = screen.getAllByText(/exclude/i)[1] // Second exclude button for GOOGL
    await user.click(excludeButton)

    // Should call preview again with excluded stocks
    await waitFor(() => {
      expect(mockGenerateAllocationPreview).toHaveBeenCalledWith(
        expect.objectContaining({
          excludedStocks: ['2'],
        })
      )
    })
  })

  it('shows loading states during async operations', async () => {
    const mockStrategies = [createMockStrategy()]
    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGenerateAllocationPreview.mockReturnValue(new Promise(() => {})) // Never resolves

    render(<PortfolioBuilder />)

    await waitFor(() => {
      expect(screen.getByLabelText(/portfolio name/i)).toBeInTheDocument()
    })

    // Fill form and generate preview
    await user.type(screen.getByLabelText(/portfolio name/i), 'Test Portfolio')
    await user.type(screen.getByLabelText(/total investment/i), '10000')
    await user.click(screen.getByRole('checkbox', { name: /test strategy/i }))
    await user.click(screen.getByText(/generate preview/i))

    // Should show loading state
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    expect(screen.getByText(/generating preview/i)).toBeInTheDocument()
  })
})