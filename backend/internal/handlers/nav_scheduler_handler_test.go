package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockNAVScheduler for testing
type MockNAVScheduler struct {
	mock.Mock
}

func (m *MockNAVScheduler) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNAVScheduler) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNAVScheduler) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNAVScheduler) GetMetrics() map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]interface{})
}

func (m *MockNAVScheduler) ForceUpdate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNAVScheduler) UpdateSinglePortfolio(portfolioID uuid.UUID) error {
	args := m.Called(portfolioID)
	return args.Error(0)
}

func setupNAVSchedulerHandler() (*fiber.App, *MockNAVScheduler) {
	app := fiber.New()
	mockScheduler := &MockNAVScheduler{}
	handler := NewNAVSchedulerHandler(mockScheduler)

	// Setup routes
	api := app.Group("/api/v1")
	api.Get("/nav-scheduler/status", handler.GetStatus)
	api.Post("/nav-scheduler/start", handler.Start)
	api.Post("/nav-scheduler/stop", handler.Stop)
	api.Post("/nav-scheduler/update", handler.ForceUpdate)
	api.Post("/nav-scheduler/update/:id", handler.UpdateSinglePortfolio)

	return app, mockScheduler
}

func TestNAVSchedulerHandler_GetStatus(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	expectedMetrics := map[string]interface{}{
		"running":           true,
		"last_update_time":  time.Now(),
		"success_count":     int64(10),
		"error_count":       int64(2),
		"total_portfolios":  5,
		"update_interval":   "15m0s",
	}

	// Setup expectations
	mockScheduler.On("GetMetrics").Return(expectedMetrics)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/nav-scheduler/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "NAV scheduler status retrieved", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, true, data["running"])
	assert.Equal(t, float64(10), data["success_count"])
	assert.Equal(t, float64(2), data["error_count"])
	assert.Equal(t, float64(5), data["total_portfolios"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_Start_Success(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("Start").Return(nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/start", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "NAV scheduler started successfully", response["message"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_Start_Error(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("Start").Return(assert.AnError)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/start", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Failed to start NAV scheduler", response["message"])
	assert.NotNil(t, response["error"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_Stop_Success(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("Stop").Return(nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/stop", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "NAV scheduler stopped successfully", response["message"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_Stop_Error(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("Stop").Return(assert.AnError)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/stop", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Failed to stop NAV scheduler", response["message"])
	assert.NotNil(t, response["error"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_ForceUpdate_Success(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("ForceUpdate").Return(nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "NAV update triggered successfully", response["message"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_ForceUpdate_Error(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectations
	mockScheduler.On("ForceUpdate").Return(assert.AnError)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Failed to trigger NAV update", response["message"])
	assert.NotNil(t, response["error"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_UpdateSinglePortfolio_Success(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	portfolioID := uuid.New()

	// Setup expectations
	mockScheduler.On("UpdateSinglePortfolio", portfolioID).Return(nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update/"+portfolioID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Portfolio NAV updated successfully", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, portfolioID.String(), data["portfolio_id"])

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_UpdateSinglePortfolio_InvalidID(t *testing.T) {
	app, _ := setupNAVSchedulerHandler()

	// Create request with invalid UUID
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update/invalid-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid portfolio ID format", response["message"])
	assert.NotNil(t, response["error"])
}

func TestNAVSchedulerHandler_UpdateSinglePortfolio_MissingID(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	// Setup expectation for the force update call that will be triggered
	mockScheduler.On("ForceUpdate").Return(assert.AnError)

	// Create request without portfolio ID - this will match the force update route
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// This actually triggers the force update endpoint, so expect 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	mockScheduler.AssertExpectations(t)
}

func TestNAVSchedulerHandler_UpdateSinglePortfolio_Error(t *testing.T) {
	app, mockScheduler := setupNAVSchedulerHandler()

	portfolioID := uuid.New()

	// Setup expectations
	mockScheduler.On("UpdateSinglePortfolio", portfolioID).Return(assert.AnError)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/nav-scheduler/update/"+portfolioID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Failed to update portfolio NAV", response["message"])
	assert.NotNil(t, response["error"])

	mockScheduler.AssertExpectations(t)
}

func TestNewNAVSchedulerHandler(t *testing.T) {
	mockScheduler := &MockNAVScheduler{}
	handler := NewNAVSchedulerHandler(mockScheduler)

	assert.NotNil(t, handler)
	assert.Equal(t, mockScheduler, handler.scheduler)
}