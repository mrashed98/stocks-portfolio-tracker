// Export all types and schemas from individual modules

// User types
export * from './user';

// Strategy types
export * from './strategy';

// Stock types
export * from './stock';

// Signal types
export * from './signal';

// Portfolio types
export * from './portfolio';

// Position types
export * from './position';

// NAV History types
export * from './nav-history';

// API types
export * from './api';

// Re-export commonly used Zod types
export { z } from 'zod';

// Common utility types
export type ID = string;
export type Timestamp = string;
export type Decimal = number;

// Form state types for React Hook Form integration
export interface FormState<T> {
  data: T;
  errors: Record<string, string>;
  isSubmitting: boolean;
  isValid: boolean;
}

// Generic CRUD operation types
export interface CrudOperations<T, CreateT, UpdateT> {
  create: (data: CreateT) => Promise<T>;
  read: (id: string) => Promise<T>;
  update: (id: string, data: UpdateT) => Promise<T>;
  delete: (id: string) => Promise<void>;
  list: (params?: any) => Promise<T[]>;
}

// Loading states for async operations
export type LoadingState = 'idle' | 'loading' | 'success' | 'error';

export interface AsyncState<T> {
  data: T | null;
  loading: LoadingState;
  error: string | null;
}

// Common component props
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

// Table column definition for data tables
export interface TableColumn<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: T) => React.ReactNode;
  width?: string;
}

// Chart configuration types
export interface ChartConfig {
  width?: number;
  height?: number;
  margin?: {
    top?: number;
    right?: number;
    bottom?: number;
    left?: number;
  };
  colors?: string[];
  theme?: 'light' | 'dark';
}

// Notification types
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  duration?: number;
  actions?: Array<{
    label: string;
    action: () => void;
  }>;
}

// Modal types
export interface ModalProps extends BaseComponentProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

// Form field types for dynamic form generation
export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'number' | 'select' | 'checkbox' | 'textarea' | 'date';
  placeholder?: string;
  required?: boolean;
  options?: Array<{ value: string; label: string }>;
  validation?: any; // Zod schema
}

// Theme types
export interface Theme {
  colors: {
    primary: string;
    secondary: string;
    success: string;
    warning: string;
    error: string;
    info: string;
    background: string;
    surface: string;
    text: string;
    textSecondary: string;
    border: string;
  };
  spacing: {
    xs: string;
    sm: string;
    md: string;
    lg: string;
    xl: string;
  };
  borderRadius: {
    sm: string;
    md: string;
    lg: string;
  };
}