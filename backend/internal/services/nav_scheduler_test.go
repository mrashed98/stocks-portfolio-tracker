package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"portfolio-app/internal/models"
)

// MockPortfolioServiceInterface for testing NAV scheduler
type MockPortfolioServiceInterface struct {
	mock.Mock
}

func (m *MockPortfolioServiceInterface) GenerateAllocationPreview(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GenerateAllocationPreviewWithExclusions(ctx context.Context, req *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error) {
	args := m.Called(ctx, req, excludedStocks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioServiceInterface) ValidateAllocationRequest(req *models.AllocationRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockPortfolioServiceInterface) CreatePortfolio(ctx context.Context, req *models.CreatePortfolioRequest, userID uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GetPortfolio(ctx context.Context, id uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GetUserPortfolios(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioServiceInterface) UpdatePortfolio(ctx context.Context, id uuid.UUID, req *models.UpdatePortfolioRequest) (*models.Portfolio, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioServiceInterface) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPortfolioServiceInterface) UpdatePortfolioNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GetPortfolioPerformanceMetrics(ctx context.Context, portfolioID uuid.UUID) (*models.PerformanceMetrics, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceMetrics), args.Error(1)
}

func (m *MockPortfolioServiceInterface) GenerateRebalancePreview(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error) {
	args := m.Called(ctx, portfolioID, newTotalInvestment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioServiceInterface) RebalancePortfolio(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.Portfolio, error) {
	args := m.Called(ctx, portfolioID, newTotalInvestment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func TestNewNAVScheduler(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	// Test with default config
	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)
	
	assert.NotNil(t, scheduler)
	assert.Equal(t, 15*time.Minute, scheduler.updateInterval)
	assert.Equal(t, 3, scheduler.maxRetries)
	assert.Equal(t, 30*time.Second, scheduler.retryDelay)
	assert.Equal(t, 10, scheduler.batchSize)
	assert.False(t, scheduler.running)
}

func TestNewNAVScheduler_WithCustomConfig(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	config := &NAVSchedulerConfig{
		UpdateInterval: 5 * time.Minute,
		MaxRetries:     5,
		RetryDelay:     1 * time.Minute,
		BatchSize:      20,
		CronExpression: "0 */5 * * * *",
	}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, config)
	
	assert.NotNil(t, scheduler)
	assert.Equal(t, 5*time.Minute, scheduler.updateInterval)
	assert.Equal(t, 5, scheduler.maxRetries)
	assert.Equal(t, 1*time.Minute, scheduler.retryDelay)
	assert.Equal(t, 20, scheduler.batchSize)
	assert.False(t, scheduler.running)
}

func TestNAVScheduler_StartStop(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	// Test start
	err := scheduler.Start()
	require.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	// Test start when already running
	err = scheduler.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = scheduler.Stop()
	require.NoError(t, err)
	assert.False(t, scheduler.IsRunning())

	// Test stop when not running
	err = scheduler.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestNAVScheduler_GetMetrics(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	metrics := scheduler.GetMetrics()
	
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "running")
	assert.Contains(t, metrics, "last_update_time")
	assert.Contains(t, metrics, "success_count")
	assert.Contains(t, metrics, "error_count")
	assert.Contains(t, metrics, "total_portfolios")
	assert.Contains(t, metrics, "update_interval")

	assert.False(t, metrics["running"].(bool))
	assert.Equal(t, int64(0), metrics["success_count"])
	assert.Equal(t, int64(0), metrics["error_count"])
	assert.Equal(t, 0, metrics["total_portfolios"])
}

func TestNAVScheduler_UpdateSinglePortfolio_Success(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)
	portfolioID := uuid.New()

	expectedNAV := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(15000.00),
		PnL:         decimal.NewFromFloat(5000.00),
		CreatedAt:   time.Now(),
	}

	// Setup expectations
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(expectedNAV, nil)

	// Execute
	err := scheduler.UpdateSinglePortfolio(portfolioID)

	// Assert
	require.NoError(t, err)
	mockPortfolioService.AssertExpectations(t)
}

func TestNAVScheduler_UpdateSinglePortfolio_WithRetry(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	// Use shorter retry delay for testing
	config := &NAVSchedulerConfig{
		UpdateInterval: 15 * time.Minute,
		MaxRetries:     2,
		RetryDelay:     10 * time.Millisecond,
		BatchSize:      10,
		CronExpression: "0 */15 * * * *",
	}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, config)
	portfolioID := uuid.New()

	expectedNAV := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(15000.00),
		PnL:         decimal.NewFromFloat(5000.00),
		CreatedAt:   time.Now(),
	}

	// Setup expectations - fail first time, succeed second time
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(nil, assert.AnError).Once()
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(expectedNAV, nil).Once()

	// Execute
	err := scheduler.UpdateSinglePortfolio(portfolioID)

	// Assert
	require.NoError(t, err)
	mockPortfolioService.AssertExpectations(t)
}

func TestNAVScheduler_UpdateSinglePortfolio_MaxRetriesExceeded(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	// Use shorter retry delay for testing
	config := &NAVSchedulerConfig{
		UpdateInterval: 15 * time.Minute,
		MaxRetries:     1,
		RetryDelay:     10 * time.Millisecond,
		BatchSize:      10,
		CronExpression: "0 */15 * * * *",
	}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, config)
	portfolioID := uuid.New()

	// Setup expectations - fail all attempts
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(nil, assert.AnError).Times(2) // maxRetries + 1

	// Execute
	err := scheduler.UpdateSinglePortfolio(portfolioID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed after")
	mockPortfolioService.AssertExpectations(t)
}

func TestNAVScheduler_ForceUpdate_NotRunning(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	// Test force update when not running
	err := scheduler.ForceUpdate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestNAVScheduler_ForceUpdate_Success(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	// Start scheduler
	require.NoError(t, scheduler.Start())
	defer scheduler.Stop()

	// Setup expectations
	portfolioIDs := []uuid.UUID{uuid.New(), uuid.New()}
	mockPortfolioRepo.On("GetAllPortfolioIDs", mock.AnythingOfType("*context.cancelCtx")).Return(portfolioIDs, nil)

	expectedNAV := &models.NAVHistory{
		PortfolioID: uuid.New(),
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(15000.00),
		PnL:         decimal.NewFromFloat(5000.00),
		CreatedAt:   time.Now(),
	}

	for _, portfolioID := range portfolioIDs {
		mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(expectedNAV, nil)
	}

	// Execute
	err := scheduler.ForceUpdate()

	// Assert
	require.NoError(t, err)

	// Give some time for the goroutine to complete
	time.Sleep(100 * time.Millisecond)

	mockPortfolioRepo.AssertExpectations(t)
	mockPortfolioService.AssertExpectations(t)
}

func TestNAVScheduler_ProcessBatches(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	config := &NAVSchedulerConfig{
		UpdateInterval: 15 * time.Minute,
		MaxRetries:     1,
		RetryDelay:     10 * time.Millisecond,
		BatchSize:      2, // Small batch size for testing
		CronExpression: "0 */15 * * * *",
	}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, config)

	// Create test portfolio IDs (more than batch size)
	portfolioIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	expectedNAV := &models.NAVHistory{
		PortfolioID: uuid.New(),
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(15000.00),
		PnL:         decimal.NewFromFloat(5000.00),
		CreatedAt:   time.Now(),
	}

	// Setup expectations for all portfolios
	for _, portfolioID := range portfolioIDs {
		mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioID).Return(expectedNAV, nil)
	}

	// Execute
	err := scheduler.processBatches(portfolioIDs)

	// Assert
	require.NoError(t, err)
	mockPortfolioService.AssertExpectations(t)

	// Check metrics
	metrics := scheduler.GetMetrics()
	assert.Equal(t, int64(5), metrics["success_count"])
	assert.Equal(t, int64(0), metrics["error_count"])
}

func TestNAVScheduler_ProcessBatches_WithErrors(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	config := &NAVSchedulerConfig{
		UpdateInterval: 15 * time.Minute,
		MaxRetries:     0, // No retries for faster test
		RetryDelay:     10 * time.Millisecond,
		BatchSize:      2,
		CronExpression: "0 */15 * * * *",
	}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, config)

	// Create test portfolio IDs
	portfolioIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	expectedNAV := &models.NAVHistory{
		PortfolioID: uuid.New(),
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(15000.00),
		PnL:         decimal.NewFromFloat(5000.00),
		CreatedAt:   time.Now(),
	}

	// Setup expectations - first succeeds, second fails, third succeeds
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioIDs[0]).Return(expectedNAV, nil)
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioIDs[1]).Return(nil, assert.AnError)
	mockPortfolioService.On("UpdatePortfolioNAV", mock.AnythingOfType("*context.cancelCtx"), portfolioIDs[2]).Return(expectedNAV, nil)

	// Execute
	err := scheduler.processBatches(portfolioIDs)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NAV update completed with")
	mockPortfolioService.AssertExpectations(t)

	// Check metrics
	metrics := scheduler.GetMetrics()
	assert.Equal(t, int64(2), metrics["success_count"])
	assert.Equal(t, int64(1), metrics["error_count"])
}

func TestDefaultNAVSchedulerConfig(t *testing.T) {
	config := DefaultNAVSchedulerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 15*time.Minute, config.UpdateInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 30*time.Second, config.RetryDelay)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, "0 */15 * * * *", config.CronExpression)
}

func TestNAVScheduler_GetAllPortfolioIDs_Error(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	// Setup expectations
	mockPortfolioRepo.On("GetAllPortfolioIDs", mock.AnythingOfType("*context.cancelCtx")).Return(nil, assert.AnError)

	// Execute
	portfolioIDs := scheduler.getAllPortfolioIDs()

	// Assert
	assert.Empty(t, portfolioIDs)
	mockPortfolioRepo.AssertExpectations(t)
}

func TestNAVScheduler_GetAllPortfolioIDs_Success(t *testing.T) {
	mockPortfolioService := &MockPortfolioServiceInterface{}
	mockPortfolioRepo := &MockPortfolioRepository{}

	scheduler := NewNAVScheduler(mockPortfolioService, mockPortfolioRepo, nil)

	expectedIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Setup expectations
	mockPortfolioRepo.On("GetAllPortfolioIDs", mock.AnythingOfType("*context.cancelCtx")).Return(expectedIDs, nil)

	// Execute
	portfolioIDs := scheduler.getAllPortfolioIDs()

	// Assert
	assert.Equal(t, expectedIDs, portfolioIDs)
	mockPortfolioRepo.AssertExpectations(t)
}