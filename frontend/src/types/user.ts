import { z } from 'zod';

// User interfaces
export interface User {
  id: string;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
}

export interface UserResponse {
  id: string;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

// Zod validation schemas
export const createUserSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be at most 255 characters'),
  email: z.string().email('Must be a valid email address').max(255, 'Email must be at most 255 characters'),
  password: z.string().min(8, 'Password must be at least 8 characters').max(128, 'Password must be at most 128 characters'),
});

export const updateUserSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be at most 255 characters').optional(),
  email: z.string().email('Must be a valid email address').max(255, 'Email must be at most 255 characters').optional(),
});

// Type inference from schemas
export type CreateUserFormData = z.infer<typeof createUserSchema>;
export type UpdateUserFormData = z.infer<typeof updateUserSchema>;