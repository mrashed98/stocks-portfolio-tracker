import axios from 'axios';
import { config } from '../lib/config';
import type { 
  Portfolio,
  PortfolioResponse,
  CreatePortfolioRequest,
  UpdatePortfolioRequest,
  AllocationPreview,
  AllocationRequest
} from '../types/portfolio';
import type { NAVHistory, PerformanceMetrics } from '../types/nav-history';

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

export interface ApiResponse<T> {
  data: T;
  cached?: boolean;
}

export interface HistoryResponse {
  data: NAVHistory[];
  from: string;
  to: string;
}

export interface RebalanceRequest {
  new_total_investment: number;
}

export const portfolioService = {
  // Generate allocation preview
  async generateAllocationPreview(request: AllocationRequest): Promise<AllocationPreview> {
    const response = await api.post<ApiResponse<AllocationPreview>>('/api/portfolios/preview', request);
    return response.data.data;
  },

  // Generate allocation preview with excluded stocks
  async generateAllocationPreviewWithExclusions(
    allocationRequest: AllocationRequest, 
    excludedStocks: string[]
  ): Promise<AllocationPreview> {
    const response = await api.post<ApiResponse<AllocationPreview>>('/api/portfolios/preview/exclude', {
      allocation_request: allocationRequest,
      excluded_stocks: excludedStocks,
    });
    return response.data.data;
  },

  // Validate allocation request
  async validateAllocationRequest(request: AllocationRequest): Promise<{ valid: boolean; error?: string }> {
    const response = await api.post<{ valid: boolean; error?: string }>('/api/portfolios/validate', request);
    return response.data;
  },

  // Create a new portfolio
  async createPortfolio(request: CreatePortfolioRequest): Promise<Portfolio> {
    const response = await api.post<ApiResponse<PortfolioResponse>>('/api/portfolios', request);
    return response.data.data;
  },

  // Get all portfolios for the current user
  async getPortfolios(): Promise<Portfolio[]> {
    const response = await api.get<ApiResponse<PortfolioResponse[]>>('/api/portfolios');
    return response.data.data;
  },

  // Get a single portfolio by ID
  async getPortfolio(id: string): Promise<Portfolio> {
    const response = await api.get<ApiResponse<PortfolioResponse>>(`/api/portfolios/${id}`);
    return response.data.data;
  },

  // Update a portfolio
  async updatePortfolio(id: string, request: UpdatePortfolioRequest): Promise<Portfolio> {
    const response = await api.put<ApiResponse<PortfolioResponse>>(`/api/portfolios/${id}`, request);
    return response.data.data;
  },

  // Delete a portfolio
  async deletePortfolio(id: string): Promise<void> {
    await api.delete(`/api/portfolios/${id}`);
  },

  // Get portfolio history
  async getPortfolioHistory(id: string, from?: string, to?: string): Promise<NAVHistory[]> {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);
    
    const response = await api.get<HistoryResponse>(`/api/portfolios/${id}/history?${params.toString()}`);
    return response.data.data;
  },

  // Get portfolio performance metrics
  async getPortfolioPerformance(id: string): Promise<PerformanceMetrics> {
    const response = await api.get<ApiResponse<PerformanceMetrics>>(`/api/portfolios/${id}/performance`);
    return response.data.data;
  },

  // Update portfolio NAV
  async updatePortfolioNAV(id: string): Promise<NAVHistory> {
    const response = await api.post<ApiResponse<NAVHistory>>(`/api/portfolios/${id}/nav/update`);
    return response.data.data;
  },

  // Generate rebalance preview
  async generateRebalancePreview(id: string, newTotalInvestment: number): Promise<AllocationPreview> {
    const response = await api.post<ApiResponse<AllocationPreview>>(
      `/api/portfolios/${id}/rebalance/preview`,
      { new_total_investment: newTotalInvestment }
    );
    return response.data.data;
  },

  // Rebalance portfolio
  async rebalancePortfolio(id: string, newTotalInvestment: number): Promise<Portfolio> {
    const response = await api.post<ApiResponse<PortfolioResponse>>(
      `/api/portfolios/${id}/rebalance`,
      { new_total_investment: newTotalInvestment }
    );
    return response.data.data;
  },

  // Clear allocation cache
  async clearAllocationCache(): Promise<void> {
    await api.delete('/api/portfolios/cache');
  },

  // NAV Scheduler endpoints
  async getNAVSchedulerStatus(): Promise<any> {
    const response = await api.get('/api/nav-scheduler/status');
    return response.data.data;
  },

  async startNAVScheduler(): Promise<void> {
    await api.post('/api/nav-scheduler/start');
  },

  async stopNAVScheduler(): Promise<void> {
    await api.post('/api/nav-scheduler/stop');
  },

  async forceNAVUpdate(): Promise<void> {
    await api.post('/api/nav-scheduler/update');
  },

  async updateSinglePortfolioNAV(portfolioId: string): Promise<void> {
    await api.post(`/api/nav-scheduler/update/${portfolioId}`);
  },
};

export default portfolioService;