import axios from 'axios';
import { config, apiEndpoints } from '../lib/config';
import type { 
  Strategy, 
  CreateStrategyRequest, 
  UpdateStrategyRequest,
  StrategyResponse 
} from '../types/strategy';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: config.apiUrl,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for authentication (when implemented)
api.interceptors.request.use((config) => {
  // TODO: Add JWT token when authentication is implemented
  // const token = localStorage.getItem('token');
  // if (token) {
  //   config.headers.Authorization = `Bearer ${token}`;
  // }
  return config;
});

// Add response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // TODO: Handle authentication errors when auth is implemented
      // localStorage.removeItem('token');
      // window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface StrategiesResponse {
  strategies: StrategyResponse[];
}

export interface UpdateStockEligibilityRequest {
  eligible: boolean;
}

export interface UpdateWeightRequest {
  weight_value: string;
}

export const strategyService = {
  // Get all strategies for the current user
  async getStrategies(): Promise<Strategy[]> {
    const response = await api.get<StrategiesResponse>(apiEndpoints.strategies.list);
    return response.data.strategies;
  },

  // Get a single strategy by ID
  async getStrategy(id: string): Promise<Strategy> {
    const response = await api.get<StrategyResponse>(`${apiEndpoints.strategies.list}/${id}`);
    return response.data;
  },

  // Create a new strategy
  async createStrategy(data: CreateStrategyRequest): Promise<Strategy> {
    const response = await api.post<StrategyResponse>(apiEndpoints.strategies.create, data);
    return response.data;
  },

  // Update an existing strategy
  async updateStrategy(id: string, data: UpdateStrategyRequest): Promise<Strategy> {
    const response = await api.put<StrategyResponse>(apiEndpoints.strategies.update(id), data);
    return response.data;
  },

  // Delete a strategy
  async deleteStrategy(id: string): Promise<void> {
    await api.delete(apiEndpoints.strategies.delete(id));
  },

  // Update strategy weight
  async updateStrategyWeight(id: string, weightValue: number): Promise<Strategy> {
    const data: UpdateWeightRequest = {
      weight_value: weightValue.toString(),
    };
    const response = await api.put<StrategyResponse>(apiEndpoints.strategies.weight(id), data);
    return response.data;
  },

  // Update stock eligibility within a strategy
  async updateStockEligibility(strategyId: string, stockId: string, eligible: boolean): Promise<void> {
    const data: UpdateStockEligibilityRequest = { eligible };
    await api.put(`${apiEndpoints.strategies.list}/${strategyId}/stocks/${stockId}`, data);
  },
};

export default strategyService;