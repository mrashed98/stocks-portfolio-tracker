import { describe, it, expect, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useToast } from '../use-toast'

describe('useToast', () => {
  it('should initialize with empty toasts', () => {
    const { result } = renderHook(() => useToast())
    
    expect(result.current.toasts).toEqual([])
  })

  it('should add a toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.toast({
        title: 'Test Toast',
        description: 'This is a test toast',
      })
    })

    expect(result.current.toasts).toHaveLength(1)
    expect(result.current.toasts[0].title).toBe('Test Toast')
    expect(result.current.toasts[0].description).toBe('This is a test toast')
  })

  it('should dismiss a toast', () => {
    const { result } = renderHook(() => useToast())
    
    let toastId: string
    
    act(() => {
      const toast = result.current.toast({
        title: 'Test Toast',
      })
      toastId = toast.id
    })

    expect(result.current.toasts).toHaveLength(1)

    act(() => {
      result.current.dismiss(toastId)
    })

    expect(result.current.toasts).toHaveLength(0)
  })

  it('should handle multiple toasts', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.toast({ title: 'Toast 1' })
      result.current.toast({ title: 'Toast 2' })
      result.current.toast({ title: 'Toast 3' })
    })

    expect(result.current.toasts).toHaveLength(3)
    expect(result.current.toasts[0].title).toBe('Toast 1')
    expect(result.current.toasts[1].title).toBe('Toast 2')
    expect(result.current.toasts[2].title).toBe('Toast 3')
  })

  it('should auto-dismiss toasts after timeout', async () => {
    vi.useFakeTimers()
    
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.toast({
        title: 'Auto-dismiss Toast',
        duration: 1000,
      })
    })

    expect(result.current.toasts).toHaveLength(1)

    // Fast-forward time
    act(() => {
      vi.advanceTimersByTime(1000)
    })

    expect(result.current.toasts).toHaveLength(0)
    
    vi.useRealTimers()
  })

  it('should handle different toast variants', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.toast({
        title: 'Success Toast',
        variant: 'default',
      })
      result.current.toast({
        title: 'Error Toast',
        variant: 'destructive',
      })
    })

    expect(result.current.toasts).toHaveLength(2)
    expect(result.current.toasts[0].variant).toBe('default')
    expect(result.current.toasts[1].variant).toBe('destructive')
  })
})