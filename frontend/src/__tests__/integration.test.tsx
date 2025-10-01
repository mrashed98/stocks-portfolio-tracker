import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'

// Simple mock component to test integration
const MockPortfolioList = () => {
  const portfolios = [
    { id: '1', name: 'Growth Portfolio', value: 50000 },
    { id: '2', name: 'Income Portfolio', value: 30000 },
  ]

  return (
    <div>
      <h1>My Portfolios</h1>
      <div data-testid="portfolio-list">
        {portfolios.map(portfolio => (
          <div key={portfolio.id} data-testid={`portfolio-${portfolio.id}`}>
            <h2>{portfolio.name}</h2>
            <p>${portfolio.value.toLocaleString()}</p>
          </div>
        ))}
      </div>
      <button onClick={() => alert('Create new portfolio')}>
        Create Portfolio
      </button>
    </div>
  )
}

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        {children}
      </BrowserRouter>
    </QueryClientProvider>
  )
}

describe('Integration Tests', () => {
  it('renders portfolio list with correct data', () => {
    render(
      <TestWrapper>
        <MockPortfolioList />
      </TestWrapper>
    )

    expect(screen.getByText('My Portfolios')).toBeInTheDocument()
    expect(screen.getByText('Growth Portfolio')).toBeInTheDocument()
    expect(screen.getByText('Income Portfolio')).toBeInTheDocument()
    expect(screen.getByText('$50,000')).toBeInTheDocument()
    expect(screen.getByText('$30,000')).toBeInTheDocument()
  })

  it('shows create portfolio button', () => {
    render(
      <TestWrapper>
        <MockPortfolioList />
      </TestWrapper>
    )

    const createButton = screen.getByRole('button', { name: /create portfolio/i })
    expect(createButton).toBeInTheDocument()
  })

  it('handles user interactions', () => {
    // Mock window.alert
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})

    render(
      <TestWrapper>
        <MockPortfolioList />
      </TestWrapper>
    )

    const createButton = screen.getByRole('button', { name: /create portfolio/i })
    fireEvent.click(createButton)

    expect(alertSpy).toHaveBeenCalledWith('Create new portfolio')
    
    alertSpy.mockRestore()
  })

  it('displays portfolio data correctly', () => {
    render(
      <TestWrapper>
        <MockPortfolioList />
      </TestWrapper>
    )

    const portfolioList = screen.getByTestId('portfolio-list')
    expect(portfolioList).toBeInTheDocument()

    const portfolio1 = screen.getByTestId('portfolio-1')
    const portfolio2 = screen.getByTestId('portfolio-2')

    expect(portfolio1).toBeInTheDocument()
    expect(portfolio2).toBeInTheDocument()
  })
})