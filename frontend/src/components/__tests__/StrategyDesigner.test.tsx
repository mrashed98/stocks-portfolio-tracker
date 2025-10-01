import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render, createMockStrategy, createMockStock, mockApiResponse } from '../../test/utils'
import StrategyDesigner from '../StrategyDesigner'

// Mock the services
const mockGetStrategies = vi.fn()
const mockCreateStrategy = vi.fn()
const mockUpdateStrategy = vi.fn()
const mockGetStocks = vi.fn()
const mockUpdateStockEligibility = vi.fn()
vi.mock('../../services/strategyService', () => ({
  getStrategies: (...args: any[]) => mockGetStrategies(...args),
  createStrategy: (...args: any[]) => mockCreateStrategy(...args),
  updateStrategy: (...args: any[]) => mockUpdateStrategy(...args),
  updateStockEligibility: (...args: any[]) => mockUpdateStockEligibility(...args),
}))

vi.mock('../../services/stockService', () => ({
  getStocks: (...args: any[]) => mockGetStocks(...args),
}))

describe('StrategyDesigner', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders strategy designer interface', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy', weightMode: 'percent', weightValue: 60 }),
      createMockStrategy({ id: '2', name: 'Value Strategy', weightMode: 'budget', weightValue: 5000 }),
    ]
    const mockStocks = [
      createMockStock({ id: '1', ticker: 'AAPL', name: 'Apple Inc.' }),
      createMockStock({ id: '2', ticker: 'GOOGL', name: 'Alphabet Inc.' }),
    ]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText('Strategy Designer')).toBeInTheDocument()
    })

    // Check existing strategies are displayed
    expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    expect(screen.getByText('Value Strategy')).toBeInTheDocument()
    expect(screen.getByText('60%')).toBeInTheDocument()
    expect(screen.getByText('$5,000')).toBeInTheDocument()

    // Check create new strategy button
    expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
  })

  it('allows user to create a new percentage-based strategy', async () => {
    const mockStrategies: any[] = []
    const mockStocks = [createMockStock()]
    const newStrategy = createMockStrategy({ id: '3', name: 'New Strategy', weightMode: 'percent', weightValue: 30 })

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))
    mockCreateStrategy.mockResolvedValue(mockApiResponse(newStrategy))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
    })

    // Click create new strategy
    await user.click(screen.getByText(/create new strategy/i))

    // Fill in strategy details
    await user.type(screen.getByLabelText(/strategy name/i), 'New Strategy')
    
    // Select percentage mode
    const percentageRadio = screen.getByLabelText(/percentage/i)
    await user.click(percentageRadio)
    
    // Enter weight value
    await user.type(screen.getByLabelText(/weight value/i), '30')

    // Submit form
    await user.click(screen.getByText(/create strategy/i))

    await waitFor(() => {
      expect(mockCreateStrategy).toHaveBeenCalledWith({
        name: 'New Strategy',
        weightMode: 'percent',
        weightValue: 30,
      })
    })
  })

  it('allows user to create a budget-based strategy', async () => {
    const mockStrategies: any[] = []
    const mockStocks = [createMockStock()]
    const newStrategy = createMockStrategy({ id: '3', name: 'Budget Strategy', weightMode: 'budget', weightValue: 10000 })

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))
    mockCreateStrategy.mockResolvedValue(mockApiResponse(newStrategy))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
    })

    // Click create new strategy
    await user.click(screen.getByText(/create new strategy/i))

    // Fill in strategy details
    await user.type(screen.getByLabelText(/strategy name/i), 'Budget Strategy')
    
    // Select budget mode
    const budgetRadio = screen.getByLabelText(/fixed budget/i)
    await user.click(budgetRadio)
    
    // Enter weight value
    await user.type(screen.getByLabelText(/weight value/i), '10000')

    // Submit form
    await user.click(screen.getByText(/create strategy/i))

    await waitFor(() => {
      expect(mockCreateStrategy).toHaveBeenCalledWith({
        name: 'Budget Strategy',
        weightMode: 'budget',
        weightValue: 10000,
      })
    })
  })

  it('validates strategy creation form', async () => {
    const mockStrategies: any[] = []
    const mockStocks = [createMockStock()]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
    })

    // Click create new strategy
    await user.click(screen.getByText(/create new strategy/i))

    // Try to submit without filling required fields
    await user.click(screen.getByText(/create strategy/i))

    await waitFor(() => {
      expect(screen.getByText(/strategy name is required/i)).toBeInTheDocument()
      expect(screen.getByText(/weight value is required/i)).toBeInTheDocument()
    })
  })

  it('validates percentage weight constraints', async () => {
    const existingStrategies = [
      createMockStrategy({ weightMode: 'percent', weightValue: 70 }),
    ]
    const mockStocks = [createMockStock()]

    mockGetStrategies.mockResolvedValue(mockApiResponse(existingStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
    })

    // Click create new strategy
    await user.click(screen.getByText(/create new strategy/i))

    // Fill form with percentage that would exceed 100%
    await user.type(screen.getByLabelText(/strategy name/i), 'Exceeding Strategy')
    await user.click(screen.getByLabelText(/percentage/i))
    await user.type(screen.getByLabelText(/weight value/i), '40') // 70% + 40% = 110%

    await user.click(screen.getByText(/create strategy/i))

    await waitFor(() => {
      expect(screen.getByText(/total percentage weights cannot exceed 100%/i)).toBeInTheDocument()
    })
  })

  it('allows user to edit existing strategy', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy', weightMode: 'percent', weightValue: 60 }),
    ]
    const mockStocks = [createMockStock()]
    const updatedStrategy = createMockStrategy({ id: '1', name: 'Updated Growth Strategy', weightValue: 50 })

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))
    mockUpdateStrategy.mockResolvedValue(mockApiResponse(updatedStrategy))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    })

    // Click edit button for the strategy
    const editButton = screen.getByRole('button', { name: /edit/i })
    await user.click(editButton)

    // Update strategy name
    const nameInput = screen.getByDisplayValue('Growth Strategy')
    await user.clear(nameInput)
    await user.type(nameInput, 'Updated Growth Strategy')

    // Update weight value
    const weightInput = screen.getByDisplayValue('60')
    await user.clear(weightInput)
    await user.type(weightInput, '50')

    // Save changes
    await user.click(screen.getByText(/save changes/i))

    await waitFor(() => {
      expect(mockUpdateStrategy).toHaveBeenCalledWith('1', {
        name: 'Updated Growth Strategy',
        weightValue: 50,
      })
    })
  })

  it('allows user to manage stock eligibility within strategies', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy' }),
    ]
    const mockStocks = [
      createMockStock({ id: '1', ticker: 'AAPL', name: 'Apple Inc.' }),
      createMockStock({ id: '2', ticker: 'GOOGL', name: 'Alphabet Inc.' }),
    ]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))
    mockUpdateStockEligibility.mockResolvedValue(mockApiResponse({}))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    })

    // Click on strategy to view stock assignments
    await user.click(screen.getByText('Growth Strategy'))

    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeInTheDocument()
      expect(screen.getByText('GOOGL')).toBeInTheDocument()
    })

    // Toggle stock eligibility
    const aaplToggle = screen.getByRole('checkbox', { name: /aapl/i })
    await user.click(aaplToggle)

    await waitFor(() => {
      expect(mockUpdateStockEligibility).toHaveBeenCalledWith('1', '1', true)
    })
  })

  it('shows real-time weight validation feedback', async () => {
    const existingStrategies = [
      createMockStrategy({ weightMode: 'percent', weightValue: 60 }),
    ]
    const mockStocks = [createMockStock()]

    mockGetStrategies.mockResolvedValue(mockApiResponse(existingStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/create new strategy/i)).toBeInTheDocument()
    })

    // Should show current total percentage
    expect(screen.getByText(/total allocated: 60%/i)).toBeInTheDocument()
    expect(screen.getByText(/remaining: 40%/i)).toBeInTheDocument()

    // Click create new strategy
    await user.click(screen.getByText(/create new strategy/i))

    // Select percentage mode and enter value
    await user.click(screen.getByLabelText(/percentage/i))
    await user.type(screen.getByLabelText(/weight value/i), '30')

    // Should show updated totals
    await waitFor(() => {
      expect(screen.getByText(/would total: 90%/i)).toBeInTheDocument()
    })
  })

  it('handles strategy deletion', async () => {
    const mockStrategies = [
      createMockStrategy({ id: '1', name: 'Growth Strategy' }),
    ]
    const mockStocks = [createMockStock()]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText('Growth Strategy')).toBeInTheDocument()
    })

    // Click delete button
    const deleteButton = screen.getByRole('button', { name: /delete/i })
    await user.click(deleteButton)

    // Confirm deletion
    await waitFor(() => {
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
    })

    const confirmButton = screen.getByText(/confirm/i)
    await user.click(confirmButton)

    // Strategy should be removed from the list
    await waitFor(() => {
      expect(screen.queryByText('Growth Strategy')).not.toBeInTheDocument()
    })
  })

  it('shows loading states during async operations', async () => {
    mockGetStrategies.mockReturnValue(new Promise(() => {})) // Never resolves
    mockGetStocks.mockResolvedValue(mockApiResponse([]))

    render(<StrategyDesigner />)

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
  })

  it('handles API errors gracefully', async () => {
    mockGetStrategies.mockRejectedValue(new Error('Failed to fetch strategies'))
    mockGetStocks.mockResolvedValue(mockApiResponse([]))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText(/failed to fetch strategies/i)).toBeInTheDocument()
    })
  })

  it('filters stocks by search term', async () => {
    const mockStrategies = [createMockStrategy()]
    const mockStocks = [
      createMockStock({ ticker: 'AAPL', name: 'Apple Inc.' }),
      createMockStock({ ticker: 'GOOGL', name: 'Alphabet Inc.' }),
      createMockStock({ ticker: 'MSFT', name: 'Microsoft Corp.' }),
    ]

    mockGetStrategies.mockResolvedValue(mockApiResponse(mockStrategies))
    mockGetStocks.mockResolvedValue(mockApiResponse(mockStocks))

    render(<StrategyDesigner />)

    await waitFor(() => {
      expect(screen.getByText('Test Strategy')).toBeInTheDocument()
    })

    // Click on strategy to view stocks
    await user.click(screen.getByText('Test Strategy'))

    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeInTheDocument()
      expect(screen.getByText('GOOGL')).toBeInTheDocument()
      expect(screen.getByText('MSFT')).toBeInTheDocument()
    })

    // Search for Apple
    const searchInput = screen.getByPlaceholderText(/search stocks/i)
    await user.type(searchInput, 'Apple')

    // Should only show Apple
    expect(screen.getByText('AAPL')).toBeInTheDocument()
    expect(screen.queryByText('GOOGL')).not.toBeInTheDocument()
    expect(screen.queryByText('MSFT')).not.toBeInTheDocument()
  })
})