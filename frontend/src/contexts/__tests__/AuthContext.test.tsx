import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '../AuthContext'
import { mockApiResponse, mockApiError, createMockUser } from '../../test/utils'

// Mock the API client
const mockLogin = vi.fn()
const mockRegister = vi.fn()
const mockGetProfile = vi.fn()
vi.mock('../../lib/api-client', () => ({
  login: (...args: any[]) => mockLogin(...args),
  register: (...args: any[]) => mockRegister(...args),
  getProfile: (...args: any[]) => mockGetProfile(...args),
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

describe('AuthContext', () => {
  let queryClient: QueryClient

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )

  beforeEach(() => {
    vi.clearAllMocks()
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })
    mockLocalStorage.getItem.mockReturnValue(null)
  })

  it('initializes with no user when no token is stored', async () => {
    mockLocalStorage.getItem.mockReturnValue(null)

    const { result } = renderHook(() => useAuth(), { wrapper })

    expect(result.current.user).toBeNull()
    expect(result.current.isLoading).toBe(false)
  })

  it('loads user profile when token exists in localStorage', async () => {
    const mockUser = createMockUser()
    const mockToken = 'valid-jwt-token'

    mockLocalStorage.getItem.mockReturnValue(mockToken)
    mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))

    const { result } = renderHook(() => useAuth(), { wrapper })

    // Initially loading
    expect(result.current.isLoading).toBe(true)

    // Wait for profile to load
    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser)
      expect(result.current.isLoading).toBe(false)
    })

    expect(mockGetProfile).toHaveBeenCalledTimes(1)
  })

  it('handles login successfully', async () => {
    const mockUser = createMockUser()
    const mockToken = 'new-jwt-token'
    const loginResponse = { user: mockUser, token: mockToken }

    mockLogin.mockResolvedValue(mockApiResponse(loginResponse))

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.login('test@example.com', 'password123')
    })

    expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123')
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('token', mockToken)
    expect(result.current.user).toEqual(mockUser)
  })

  it('handles login failure', async () => {
    const errorMessage = 'Invalid credentials'
    mockLogin.mockRejectedValue(mockApiError(errorMessage, 401))

    const { result } = renderHook(() => useAuth(), { wrapper })

    await expect(
      act(async () => {
        await result.current.login('test@example.com', 'wrongpassword')
      })
    ).rejects.toThrow(errorMessage)

    expect(result.current.user).toBeNull()
    expect(mockLocalStorage.setItem).not.toHaveBeenCalled()
  })

  it('handles registration successfully', async () => {
    const mockUser = createMockUser()
    const mockToken = 'new-jwt-token'
    const registerResponse = { user: mockUser, token: mockToken }

    mockRegister.mockResolvedValue(mockApiResponse(registerResponse))

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.register('John Doe', 'test@example.com', 'password123')
    })

    expect(mockRegister).toHaveBeenCalledWith('John Doe', 'test@example.com', 'password123')
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith('token', mockToken)
    expect(result.current.user).toEqual(mockUser)
  })

  it('handles registration failure', async () => {
    const errorMessage = 'Email already exists'
    mockRegister.mockRejectedValue(mockApiError(errorMessage, 409))

    const { result } = renderHook(() => useAuth(), { wrapper })

    await expect(
      act(async () => {
        await result.current.register('John Doe', 'existing@example.com', 'password123')
      })
    ).rejects.toThrow(errorMessage)

    expect(result.current.user).toBeNull()
    expect(mockLocalStorage.setItem).not.toHaveBeenCalled()
  })

  it('handles logout', async () => {
    const mockUser = createMockUser()
    const mockToken = 'valid-jwt-token'

    // Setup initial authenticated state
    mockLocalStorage.getItem.mockReturnValue(mockToken)
    mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))

    const { result } = renderHook(() => useAuth(), { wrapper })

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser)
    })

    // Logout
    act(() => {
      result.current.logout()
    })

    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('token')
    expect(result.current.user).toBeNull()
  })

  it('handles profile fetch failure on initialization', async () => {
    const mockToken = 'invalid-jwt-token'
    mockLocalStorage.getItem.mockReturnValue(mockToken)
    mockGetProfile.mockRejectedValue(mockApiError('Unauthorized', 401))

    const { result } = renderHook(() => useAuth(), { wrapper })

    await waitFor(() => {
      expect(result.current.user).toBeNull()
      expect(result.current.isLoading).toBe(false)
    })

    // Should remove invalid token
    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('token')
  })

  it('provides loading state during authentication operations', async () => {
    const mockUser = createMockUser()
    const mockToken = 'new-jwt-token'
    const loginResponse = { user: mockUser, token: mockToken }

    // Make login take some time
    mockLogin.mockImplementation(() => 
      new Promise(resolve => setTimeout(() => resolve(mockApiResponse(loginResponse)), 100))
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    // Start login
    const loginPromise = act(async () => {
      await result.current.login('test@example.com', 'password123')
    })

    // Should be loading
    expect(result.current.isLoading).toBe(true)

    // Wait for completion
    await loginPromise

    expect(result.current.isLoading).toBe(false)
    expect(result.current.user).toEqual(mockUser)
  })

  it('persists authentication state across hook re-renders', async () => {
    const mockUser = createMockUser()
    const mockToken = 'valid-jwt-token'

    mockLocalStorage.getItem.mockReturnValue(mockToken)
    mockGetProfile.mockResolvedValue(mockApiResponse(mockUser))

    const { result, rerender } = renderHook(() => useAuth(), { wrapper })

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser)
    })

    // Re-render the hook
    rerender()

    // User should still be authenticated
    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isLoading).toBe(false)
  })

  it('handles concurrent authentication operations', async () => {
    const mockUser = createMockUser()
    const mockToken = 'new-jwt-token'
    const loginResponse = { user: mockUser, token: mockToken }

    mockLogin.mockResolvedValue(mockApiResponse(loginResponse))

    const { result } = renderHook(() => useAuth(), { wrapper })

    // Start multiple login attempts
    const login1 = act(async () => {
      await result.current.login('test@example.com', 'password123')
    })

    const login2 = act(async () => {
      await result.current.login('test@example.com', 'password123')
    })

    await Promise.all([login1, login2])

    // Should only call login once due to React Query deduplication
    expect(mockLogin).toHaveBeenCalledTimes(1)
    expect(result.current.user).toEqual(mockUser)
  })

  it('clears user data on authentication error', async () => {
    const mockUser = createMockUser()
    const mockToken = 'valid-jwt-token'

    // Setup initial authenticated state
    mockLocalStorage.getItem.mockReturnValue(mockToken)
    mockGetProfile.mockResolvedValueOnce(mockApiResponse(mockUser))

    const { result } = renderHook(() => useAuth(), { wrapper })

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser)
    })

    // Simulate token expiration on subsequent call
    mockGetProfile.mockRejectedValue(mockApiError('Token expired', 401))

    // Trigger a profile refresh (this might happen on app focus)
    await act(async () => {
      queryClient.invalidateQueries({ queryKey: ['profile'] })
    })

    await waitFor(() => {
      expect(result.current.user).toBeNull()
    })

    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('token')
  })

  it('validates email format during registration', async () => {
    const { result } = renderHook(() => useAuth(), { wrapper })

    await expect(
      act(async () => {
        await result.current.register('John Doe', 'invalid-email', 'password123')
      })
    ).rejects.toThrow(/invalid email/i)

    expect(mockRegister).not.toHaveBeenCalled()
  })

  it('validates password strength during registration', async () => {
    const { result } = renderHook(() => useAuth(), { wrapper })

    await expect(
      act(async () => {
        await result.current.register('John Doe', 'test@example.com', '123')
      })
    ).rejects.toThrow(/password too short/i)

    expect(mockRegister).not.toHaveBeenCalled()
  })
})