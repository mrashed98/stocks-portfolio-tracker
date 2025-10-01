package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"portfolio-app/internal/models"
)

// PortfolioRepository handles portfolio data access
type PortfolioRepository struct {
	db *sql.DB
}

// PortfolioRepositoryInterface defines the portfolio repository contract
type PortfolioRepositoryInterface interface {
	Create(ctx context.Context, portfolio *models.Portfolio) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Portfolio, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error)
	GetAllPortfolioIDs(ctx context.Context) ([]uuid.UUID, error)
	Update(ctx context.Context, portfolio *models.Portfolio) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Position operations
	CreatePosition(ctx context.Context, position *models.Position) error
	GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]*models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	DeletePosition(ctx context.Context, portfolioID, stockID uuid.UUID) error
	
	// NAV History operations
	CreateNAVHistory(ctx context.Context, navHistory *models.NAVHistory) error
	GetNAVHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error)
	GetLatestNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error)
	
	// Batch operations
	CreatePortfolioWithPositions(ctx context.Context, portfolio *models.Portfolio, positions []*models.Position) error
}

// NewPortfolioRepository creates a new portfolio repository
func NewPortfolioRepository(db *sql.DB) *PortfolioRepository {
	return &PortfolioRepository{
		db: db,
	}
}

// Create creates a new portfolio
func (r *PortfolioRepository) Create(ctx context.Context, portfolio *models.Portfolio) error {
	query := `
		INSERT INTO portfolios (id, user_id, name, total_investment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.ExecContext(ctx, query, portfolio.ID, portfolio.UserID, portfolio.Name, 
		portfolio.TotalInvestment, portfolio.CreatedAt, portfolio.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create portfolio: %w", err)
	}
	
	return nil
}

// GetByID retrieves a portfolio by ID with positions and latest NAV
func (r *PortfolioRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Portfolio, error) {
	portfolio := &models.Portfolio{}
	
	query := `
		SELECT id, user_id, name, total_investment, created_at, updated_at
		FROM portfolios 
		WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.TotalInvestment,
		&portfolio.CreatedAt, &portfolio.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("portfolio not found")
		}
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	// Load positions
	positions, err := r.GetPositions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}
	portfolio.Positions = make([]models.Position, len(positions))
	for i, pos := range positions {
		portfolio.Positions[i] = *pos
	}
	
	// Load latest NAV
	latestNAV, err := r.GetLatestNAV(ctx, id)
	if err == nil && latestNAV != nil {
		portfolio.NAVHistory = []models.NAVHistory{*latestNAV}
	}
	
	return portfolio, nil
}

// GetByUserID retrieves all portfolios for a user
func (r *PortfolioRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error) {
	var portfolios []*models.Portfolio
	
	query := `
		SELECT id, user_id, name, total_investment, created_at, updated_at
		FROM portfolios 
		WHERE user_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolios: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		portfolio := &models.Portfolio{}
		err := rows.Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, 
			&portfolio.TotalInvestment, &portfolio.CreatedAt, &portfolio.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan portfolio: %w", err)
		}
		portfolios = append(portfolios, portfolio)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating portfolios: %w", err)
	}
	
	// Load positions and latest NAV for each portfolio
	for _, portfolio := range portfolios {
		positions, err := r.GetPositions(ctx, portfolio.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load positions for portfolio %s: %w", portfolio.ID, err)
		}
		portfolio.Positions = make([]models.Position, len(positions))
		for i, pos := range positions {
			portfolio.Positions[i] = *pos
		}
		
		latestNAV, err := r.GetLatestNAV(ctx, portfolio.ID)
		if err == nil && latestNAV != nil {
			portfolio.NAVHistory = []models.NAVHistory{*latestNAV}
		}
	}
	
	return portfolios, nil
}

// GetAllPortfolioIDs retrieves all portfolio IDs for NAV updates
func (r *PortfolioRepository) GetAllPortfolioIDs(ctx context.Context) ([]uuid.UUID, error) {
	var portfolioIDs []uuid.UUID
	
	query := `SELECT id FROM portfolios ORDER BY created_at`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio IDs: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan portfolio ID: %w", err)
		}
		portfolioIDs = append(portfolioIDs, id)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating portfolio IDs: %w", err)
	}
	
	return portfolioIDs, nil
}

// Update updates a portfolio
func (r *PortfolioRepository) Update(ctx context.Context, portfolio *models.Portfolio) error {
	query := `
		UPDATE portfolios 
		SET name = $1, total_investment = $2, updated_at = $3
		WHERE id = $4`
	
	result, err := r.db.ExecContext(ctx, query, portfolio.Name, portfolio.TotalInvestment, 
		portfolio.UpdatedAt, portfolio.ID)
	if err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("portfolio not found")
	}
	
	return nil
}

// Delete deletes a portfolio
func (r *PortfolioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM portfolios WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("portfolio not found")
	}
	
	return nil
}

// CreatePosition creates a new position
func (r *PortfolioRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	query := `
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.ExecContext(ctx, query, position.PortfolioID, position.StockID, position.Quantity,
		position.EntryPrice, position.AllocationValue, position.StrategyContrib,
		position.CreatedAt, position.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}
	
	return nil
}

// GetPositions retrieves all positions for a portfolio
func (r *PortfolioRepository) GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]*models.Position, error) {
	var positions []*models.Position
	
	query := `
		SELECT p.portfolio_id, p.stock_id, p.quantity, p.entry_price, p.allocation_value, 
		       p.strategy_contrib, p.created_at, p.updated_at,
		       s.ticker, s.name, s.sector
		FROM positions p
		JOIN stocks s ON p.stock_id = s.id
		WHERE p.portfolio_id = $1
		ORDER BY p.created_at`
	
	rows, err := r.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		position := &models.Position{}
		stock := &models.Stock{}
		
		err := rows.Scan(
			&position.PortfolioID, &position.StockID, &position.Quantity,
			&position.EntryPrice, &position.AllocationValue, &position.StrategyContrib,
			&position.CreatedAt, &position.UpdatedAt,
			&stock.Ticker, &stock.Name, &stock.Sector,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		
		stock.ID = position.StockID
		position.Stock = stock
		
		// Parse strategy contribution JSON
		if len(position.StrategyContrib) > 0 {
			var contribMap map[string]decimal.Decimal
			if err := json.Unmarshal(position.StrategyContrib, &contribMap); err == nil {
				position.StrategyContribMap = contribMap
			}
		}
		
		positions = append(positions, position)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}
	
	return positions, nil
}

// UpdatePosition updates a position
func (r *PortfolioRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	query := `
		UPDATE positions 
		SET quantity = $1, entry_price = $2, allocation_value = $3, 
		    strategy_contrib = $4, updated_at = $5
		WHERE portfolio_id = $6 AND stock_id = $7`
	
	result, err := r.db.ExecContext(ctx, query, position.Quantity, position.EntryPrice, 
		position.AllocationValue, position.StrategyContrib, position.UpdatedAt,
		position.PortfolioID, position.StockID)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("position not found")
	}
	
	return nil
}

// DeletePosition deletes a position
func (r *PortfolioRepository) DeletePosition(ctx context.Context, portfolioID, stockID uuid.UUID) error {
	query := `DELETE FROM positions WHERE portfolio_id = $1 AND stock_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, portfolioID, stockID)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("position not found")
	}
	
	return nil
}

// CreateNAVHistory creates a new NAV history entry
func (r *PortfolioRepository) CreateNAVHistory(ctx context.Context, navHistory *models.NAVHistory) error {
	query := `
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.ExecContext(ctx, query, navHistory.PortfolioID, navHistory.Timestamp, 
		navHistory.NAV, navHistory.PnL, navHistory.Drawdown, navHistory.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create NAV history: %w", err)
	}
	
	return nil
}

// GetNAVHistory retrieves NAV history for a portfolio within a date range
func (r *PortfolioRepository) GetNAVHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error) {
	var navHistory []*models.NAVHistory
	
	query := `
		SELECT portfolio_id, timestamp, nav, pnl, drawdown, created_at
		FROM nav_history 
		WHERE portfolio_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC`
	
	rows, err := r.db.QueryContext(ctx, query, portfolioID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAV history: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		nav := &models.NAVHistory{}
		err := rows.Scan(&nav.PortfolioID, &nav.Timestamp, &nav.NAV, &nav.PnL, 
			&nav.Drawdown, &nav.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan NAV history: %w", err)
		}
		navHistory = append(navHistory, nav)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating NAV history: %w", err)
	}
	
	return navHistory, nil
}

// GetLatestNAV retrieves the latest NAV entry for a portfolio
func (r *PortfolioRepository) GetLatestNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error) {
	navHistory := &models.NAVHistory{}
	
	query := `
		SELECT portfolio_id, timestamp, nav, pnl, drawdown, created_at
		FROM nav_history 
		WHERE portfolio_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`
	
	err := r.db.QueryRowContext(ctx, query, portfolioID).Scan(
		&navHistory.PortfolioID, &navHistory.Timestamp, &navHistory.NAV, 
		&navHistory.PnL, &navHistory.Drawdown, &navHistory.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No NAV history yet
		}
		return nil, fmt.Errorf("failed to get latest NAV: %w", err)
	}
	
	return navHistory, nil
}

// CreatePortfolioWithPositions creates a portfolio and its positions in a transaction
func (r *PortfolioRepository) CreatePortfolioWithPositions(ctx context.Context, portfolio *models.Portfolio, positions []*models.Position) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Create portfolio
	portfolioQuery := `
		INSERT INTO portfolios (id, user_id, name, total_investment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err = tx.ExecContext(ctx, portfolioQuery,
		portfolio.ID, portfolio.UserID, portfolio.Name, portfolio.TotalInvestment,
		portfolio.CreatedAt, portfolio.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create portfolio: %w", err)
	}
	
	// Create positions
	if len(positions) > 0 {
		positionQuery := `
			INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		
		for _, position := range positions {
			_, err = tx.ExecContext(ctx, positionQuery,
				position.PortfolioID, position.StockID, position.Quantity,
				position.EntryPrice, position.AllocationValue, position.StrategyContrib,
				position.CreatedAt, position.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create position for stock %s: %w", position.StockID, err)
			}
		}
	}
	
	// Create initial NAV history entry
	initialNAV := &models.NAVHistory{
		PortfolioID: portfolio.ID,
		Timestamp:   portfolio.CreatedAt,
		NAV:         portfolio.TotalInvestment,
		PnL:         decimal.Zero,
		Drawdown:    nil,
		CreatedAt:   portfolio.CreatedAt,
	}
	
	navQuery := `
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err = tx.ExecContext(ctx, navQuery,
		initialNAV.PortfolioID, initialNAV.Timestamp, initialNAV.NAV,
		initialNAV.PnL, initialNAV.Drawdown, initialNAV.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create initial NAV history: %w", err)
	}
	
	return tx.Commit()
}