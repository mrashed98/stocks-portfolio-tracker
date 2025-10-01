package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// NAVScheduler handles background NAV updates for portfolios
type NAVScheduler struct {
	portfolioService PortfolioServiceInterface
	portfolioRepo    PortfolioRepository
	cron             *cron.Cron
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	running          bool
	mu               sync.RWMutex
	
	// Configuration
	updateInterval   time.Duration
	maxRetries       int
	retryDelay       time.Duration
	batchSize        int
	
	// Metrics
	lastUpdateTime   time.Time
	successCount     int64
	errorCount       int64
	totalPortfolios  int
}

// NAVSchedulerConfig holds configuration for the NAV scheduler
type NAVSchedulerConfig struct {
	UpdateInterval  time.Duration // How often to update NAV (default: 15 minutes)
	MaxRetries      int           // Maximum retries for failed updates (default: 3)
	RetryDelay      time.Duration // Delay between retries (default: 30 seconds)
	BatchSize       int           // Number of portfolios to process in parallel (default: 10)
	CronExpression  string        // Cron expression for scheduling (default: "*/15 * * * *")
}

// DefaultNAVSchedulerConfig returns default configuration
func DefaultNAVSchedulerConfig() *NAVSchedulerConfig {
	return &NAVSchedulerConfig{
		UpdateInterval: 15 * time.Minute,
		MaxRetries:     3,
		RetryDelay:     30 * time.Second,
		BatchSize:      10,
		CronExpression: "0 */15 * * * *", // Every 15 minutes (with seconds field)
	}
}

// NewNAVScheduler creates a new NAV scheduler
func NewNAVScheduler(portfolioService PortfolioServiceInterface, portfolioRepo PortfolioRepository, config *NAVSchedulerConfig) *NAVScheduler {
	if config == nil {
		config = DefaultNAVSchedulerConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	scheduler := &NAVScheduler{
		portfolioService: portfolioService,
		portfolioRepo:    portfolioRepo,
		cron:             cron.New(cron.WithSeconds()),
		ctx:              ctx,
		cancel:           cancel,
		updateInterval:   config.UpdateInterval,
		maxRetries:       config.MaxRetries,
		retryDelay:       config.RetryDelay,
		batchSize:        config.BatchSize,
	}
	
	// Add cron job for NAV updates
	_, err := scheduler.cron.AddFunc(config.CronExpression, scheduler.scheduleNAVUpdate)
	if err != nil {
		log.Printf("Failed to add NAV update cron job: %v", err)
	}
	
	return scheduler
}

// Start begins the NAV scheduler
func (s *NAVScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return fmt.Errorf("NAV scheduler is already running")
	}
	
	log.Println("Starting NAV scheduler...")
	s.cron.Start()
	s.running = true
	
	// Run initial update
	go s.scheduleNAVUpdate()
	
	log.Printf("NAV scheduler started with %d minute intervals", int(s.updateInterval.Minutes()))
	return nil
}

// Stop gracefully stops the NAV scheduler
func (s *NAVScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return fmt.Errorf("NAV scheduler is not running")
	}
	
	log.Println("Stopping NAV scheduler...")
	
	// Stop cron scheduler
	s.cron.Stop()
	
	// Cancel context to stop background operations
	s.cancel()
	
	// Wait for all goroutines to finish
	s.wg.Wait()
	
	s.running = false
	log.Println("NAV scheduler stopped")
	
	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *NAVScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetMetrics returns scheduler metrics
func (s *NAVScheduler) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return map[string]interface{}{
		"running":           s.running,
		"last_update_time":  s.lastUpdateTime,
		"success_count":     s.successCount,
		"error_count":       s.errorCount,
		"total_portfolios":  s.totalPortfolios,
		"update_interval":   s.updateInterval.String(),
	}
}

// scheduleNAVUpdate is called by the cron scheduler to update all portfolio NAVs
func (s *NAVScheduler) scheduleNAVUpdate() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	
	log.Println("Starting scheduled NAV update for all portfolios")
	
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		start := time.Now()
		err := s.updateAllPortfolioNAVs()
		duration := time.Since(start)
		
		s.mu.Lock()
		s.lastUpdateTime = time.Now()
		s.mu.Unlock()
		
		if err != nil {
			log.Printf("NAV update completed with errors in %v: %v", duration, err)
		} else {
			log.Printf("NAV update completed successfully in %v", duration)
		}
	}()
}

// updateAllPortfolioNAVs updates NAV for all portfolios
func (s *NAVScheduler) updateAllPortfolioNAVs() error {
	// Get all portfolios - this would require a method to get all portfolio IDs
	// For now, we'll implement a simple approach that requires portfolio repository integration
	
	// Since we don't have a direct way to get all portfolio IDs from the service,
	// we'll need to add this functionality. For now, let's create a placeholder
	// that can be extended when portfolio repository is integrated.
	
	log.Println("NAV update: Getting list of all portfolios...")
	
	// This is a placeholder - in a real implementation, you would:
	// 1. Get all portfolio IDs from the database
	// 2. Process them in batches
	// 3. Update NAV for each portfolio
	
	portfolioIDs := s.getAllPortfolioIDs()
	if len(portfolioIDs) == 0 {
		log.Println("No portfolios found for NAV update")
		return nil
	}
	
	s.mu.Lock()
	s.totalPortfolios = len(portfolioIDs)
	s.mu.Unlock()
	
	log.Printf("Updating NAV for %d portfolios", len(portfolioIDs))
	
	// Process portfolios in batches
	return s.processBatches(portfolioIDs)
}

// getAllPortfolioIDs gets all portfolio IDs that need NAV updates
func (s *NAVScheduler) getAllPortfolioIDs() []uuid.UUID {
	portfolioIDs, err := s.portfolioRepo.GetAllPortfolioIDs(s.ctx)
	if err != nil {
		log.Printf("Failed to get portfolio IDs for NAV update: %v", err)
		return []uuid.UUID{}
	}
	return portfolioIDs
}

// processBatches processes portfolio NAV updates in batches
func (s *NAVScheduler) processBatches(portfolioIDs []uuid.UUID) error {
	var allErrors []error
	
	// Process in batches to avoid overwhelming the system
	for i := 0; i < len(portfolioIDs); i += s.batchSize {
		end := i + s.batchSize
		if end > len(portfolioIDs) {
			end = len(portfolioIDs)
		}
		
		batch := portfolioIDs[i:end]
		
		// Check if context is cancelled
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("NAV update cancelled")
		default:
		}
		
		log.Printf("Processing NAV update batch %d-%d of %d portfolios", i+1, end, len(portfolioIDs))
		
		// Process batch in parallel
		batchErrors := s.processBatch(batch)
		if len(batchErrors) > 0 {
			allErrors = append(allErrors, batchErrors...)
		}
		
		// Small delay between batches to avoid overwhelming external APIs
		if end < len(portfolioIDs) {
			time.Sleep(1 * time.Second)
		}
	}
	
	if len(allErrors) > 0 {
		return fmt.Errorf("NAV update completed with %d errors: %v", len(allErrors), allErrors[0])
	}
	
	return nil
}

// processBatch processes a batch of portfolios in parallel
func (s *NAVScheduler) processBatch(portfolioIDs []uuid.UUID) []error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	
	for _, portfolioID := range portfolioIDs {
		wg.Add(1)
		
		go func(id uuid.UUID) {
			defer wg.Done()
			
			err := s.updatePortfolioNAVWithRetry(id)
			
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("portfolio %s: %w", id, err))
				s.errorCount++
			} else {
				s.successCount++
			}
			mu.Unlock()
		}(portfolioID)
	}
	
	wg.Wait()
	return errors
}

// updatePortfolioNAVWithRetry updates a single portfolio's NAV with retry logic
func (s *NAVScheduler) updatePortfolioNAVWithRetry(portfolioID uuid.UUID) error {
	var lastErr error
	
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("update cancelled")
		default:
		}
		
		// Attempt to update NAV
		_, err := s.portfolioService.UpdatePortfolioNAV(s.ctx, portfolioID)
		if err == nil {
			if attempt > 0 {
				log.Printf("Portfolio %s NAV updated successfully after %d retries", portfolioID, attempt)
			}
			return nil
		}
		
		lastErr = err
		
		// Don't retry on the last attempt
		if attempt < s.maxRetries {
			log.Printf("Portfolio %s NAV update failed (attempt %d/%d): %v. Retrying in %v...", 
				portfolioID, attempt+1, s.maxRetries+1, err, s.retryDelay)
			
			// Wait before retry
			select {
			case <-s.ctx.Done():
				return fmt.Errorf("update cancelled during retry")
			case <-time.After(s.retryDelay):
			}
		}
	}
	
	log.Printf("Portfolio %s NAV update failed after %d attempts: %v", portfolioID, s.maxRetries+1, lastErr)
	return fmt.Errorf("failed after %d attempts: %w", s.maxRetries+1, lastErr)
}

// ForceUpdate triggers an immediate NAV update for all portfolios
func (s *NAVScheduler) ForceUpdate() error {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.mu.RUnlock()
	
	log.Println("Force updating NAV for all portfolios...")
	
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		start := time.Now()
		err := s.updateAllPortfolioNAVs()
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("Force NAV update completed with errors in %v: %v", duration, err)
		} else {
			log.Printf("Force NAV update completed successfully in %v", duration)
		}
	}()
	
	return nil
}

// UpdateSinglePortfolio updates NAV for a specific portfolio
func (s *NAVScheduler) UpdateSinglePortfolio(portfolioID uuid.UUID) error {
	log.Printf("Updating NAV for portfolio %s", portfolioID)
	
	err := s.updatePortfolioNAVWithRetry(portfolioID)
	if err != nil {
		log.Printf("Failed to update NAV for portfolio %s: %v", portfolioID, err)
		return err
	}
	
	log.Printf("Successfully updated NAV for portfolio %s", portfolioID)
	return nil
}