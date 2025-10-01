import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { render, createMockUser } from '../../test/utils'
import Navigation from '../Navigation'
import * as AuthContext from '../../contexts/AuthContext'

// Mock the AuthContext
const mockUseAuth = vi.fn()
vi.mock('../../contexts/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}))

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => ({ pathname: '/dashboard' }),
  }
})

describe('Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders navigation links when user is authenticated', () => {
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    expect(screen.getByText('Portfolio Manager')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Portfolios')).toBeInTheDocument()
    expect(screen.getByText('Strategies')).toBeInTheDocument()
    expect(screen.getByText('Stocks')).toBeInTheDocument()
    expect(screen.getByText('Charts')).toBeInTheDocument()
  })

  it('shows user menu when authenticated', () => {
    const mockUser = createMockUser({ name: 'John Doe' })
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('calls logout when logout button is clicked', async () => {
    const mockLogout = vi.fn()
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: mockLogout,
      isLoading: false,
    })

    render(<Navigation />)

    // Click on user menu to open dropdown
    fireEvent.click(screen.getByText(mockUser.name))

    // Wait for dropdown to appear and click logout
    await waitFor(() => {
      const logoutButton = screen.getByText('Logout')
      expect(logoutButton).toBeInTheDocument()
      fireEvent.click(logoutButton)
    })

    expect(mockLogout).toHaveBeenCalledTimes(1)
  })

  it('does not render navigation when user is not authenticated', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
    expect(screen.queryByText('Portfolios')).not.toBeInTheDocument()
  })

  it('shows loading state', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      logout: vi.fn(),
      isLoading: true,
    })

    render(<Navigation />)

    // Should not render navigation items while loading
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()
  })

  it('highlights active navigation item', () => {
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    const dashboardLink = screen.getByText('Dashboard').closest('a')
    expect(dashboardLink).toHaveClass('bg-gray-900') // Active state class
  })

  it('navigates to correct routes when navigation items are clicked', () => {
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    // Test portfolio navigation
    const portfoliosLink = screen.getByText('Portfolios')
    expect(portfoliosLink.closest('a')).toHaveAttribute('href', '/portfolios')

    // Test strategies navigation
    const strategiesLink = screen.getByText('Strategies')
    expect(strategiesLink.closest('a')).toHaveAttribute('href', '/strategies')

    // Test stocks navigation
    const stocksLink = screen.getByText('Stocks')
    expect(stocksLink.closest('a')).toHaveAttribute('href', '/stocks')

    // Test charts navigation
    const chartsLink = screen.getByText('Charts')
    expect(chartsLink.closest('a')).toHaveAttribute('href', '/charts')
  })

  it('renders mobile menu toggle', () => {
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    // Mobile menu button should be present (usually hidden on desktop)
    const mobileMenuButton = screen.getByRole('button', { name: /menu/i })
    expect(mobileMenuButton).toBeInTheDocument()
  })

  it('toggles mobile menu when button is clicked', async () => {
    const mockUser = createMockUser()
    mockUseAuth.mockReturnValue({
      user: mockUser,
      logout: vi.fn(),
      isLoading: false,
    })

    render(<Navigation />)

    const mobileMenuButton = screen.getByRole('button', { name: /menu/i })
    
    // Click to open mobile menu
    fireEvent.click(mobileMenuButton)

    // Mobile menu should be visible
    await waitFor(() => {
      expect(screen.getByTestId('mobile-menu')).toBeInTheDocument()
    })

    // Click again to close
    fireEvent.click(mobileMenuButton)

    await waitFor(() => {
      expect(screen.queryByTestId('mobile-menu')).not.toBeInTheDocument()
    })
  })
})