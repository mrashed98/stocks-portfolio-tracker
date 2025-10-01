package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"portfolio-app/internal/models"
)

// StrategyRepository defines the interface for strategy data operations
type StrategyRepository interface {
	Create(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error)
	Update(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error
	RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error
	UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool) error
	GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error)
}

// strategyRepository implements the StrategyRepository interface
type strategyRepository struct {
	db *sql.DB
}

// NewStrategyRepository creates a new strategy repository instance
func NewStrategyRepository(db *sql.DB) StrategyRepository {
	return &strategyRepository{db: db}
}

// Create creates a new strategy in the database
func (r *strategyRepository) Create(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error) {
	query := `
		INSERT INTO strategies (id, user_id, name, weight_mode, weight_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, weight_mode, weight_value, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		strategy.ID,
		strategy.UserID,
		strategy.Name,
		strategy.WeightMode,
		strategy.WeightValue,
		strategy.CreatedAt,
		strategy.UpdatedAt,
	)

	var created models.Strategy
	err := row.Scan(
		&created.ID,
		&created.UserID,
		&created.Name,
		&created.WeightMode,
		&created.WeightValue,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	return &created, nil
}

// Update updates an existing strategy in the database
func (r *strategyRepository) Update(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error) {
	query := `
		UPDATE strategies 
		SET name = $3, weight_mode = $4, weight_value = $5, updated_at = $6
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, weight_mode, weight_value, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		strategy.ID,
		strategy.UserID,
		strategy.Name,
		strategy.WeightMode,
		strategy.WeightValue,
		strategy.UpdatedAt,
	)

	var updated models.Strategy
	err := row.Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Name,
		&updated.WeightMode,
		&updated.WeightValue,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "strategy"}
		}
		return nil, fmt.Errorf("failed to update strategy: %w", err)
	}

	return &updated, nil
}

// GetByID retrieves a strategy by ID and user ID
func (r *strategyRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error) {
	query := `
		SELECT id, user_id, name, weight_mode, weight_value, created_at, updated_at
		FROM strategies
		WHERE id = $1 AND user_id = $2`

	row := r.db.QueryRowContext(ctx, query, id, userID)

	var strategy models.Strategy
	err := row.Scan(
		&strategy.ID,
		&strategy.UserID,
		&strategy.Name,
		&strategy.WeightMode,
		&strategy.WeightValue,
		&strategy.CreatedAt,
		&strategy.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "strategy"}
		}
		return nil, fmt.Errorf("failed to get strategy: %w", err)
	}

	// Load associated stocks
	stocks, err := r.GetStrategyStocks(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load strategy stocks: %w", err)
	}
	// Convert []*models.StrategyStock to []models.StrategyStock
	strategy.Stocks = make([]models.StrategyStock, len(stocks))
	for i, stock := range stocks {
		strategy.Stocks[i] = *stock
	}

	return &strategy, nil
}

// GetByIDs retrieves multiple strategies by their IDs
func (r *strategyRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error) {
	if len(ids) == 0 {
		return []*models.Strategy{}, nil
	}

	var strategies []*models.Strategy
	
	// For simplicity, query each strategy individually
	// In a production system, you might want to optimize this with a single query
	for _, id := range ids {
		// Query without user_id constraint since this is used by allocation engine
		query := `
			SELECT id, user_id, name, weight_mode, weight_value, created_at, updated_at
			FROM strategies
			WHERE id = $1`

		row := r.db.QueryRowContext(ctx, query, id)

		strategy := &models.Strategy{}
		err := row.Scan(
			&strategy.ID,
			&strategy.UserID,
			&strategy.Name,
			&strategy.WeightMode,
			&strategy.WeightValue,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				continue // Skip missing strategies
			}
			return nil, fmt.Errorf("failed to get strategy %s: %w", id, err)
		}

		// Load associated stocks for each strategy
		stocks, err := r.GetStrategyStocks(ctx, strategy.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load strategy stocks for strategy %s: %w", strategy.ID, err)
		}
		// Convert []*models.StrategyStock to []models.StrategyStock
		strategy.Stocks = make([]models.StrategyStock, len(stocks))
		for i, stock := range stocks {
			strategy.Stocks[i] = *stock
		}

		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

// GetByUserID retrieves all strategies for a user
func (r *strategyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error) {
	query := `
		SELECT id, user_id, name, weight_mode, weight_value, created_at, updated_at
		FROM strategies
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query strategies: %w", err)
	}
	defer rows.Close()

	var strategies []*models.Strategy
	for rows.Next() {
		var strategy models.Strategy
		err := rows.Scan(
			&strategy.ID,
			&strategy.UserID,
			&strategy.Name,
			&strategy.WeightMode,
			&strategy.WeightValue,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan strategy: %w", err)
		}

		// Load associated stocks for each strategy
		stocks, err := r.GetStrategyStocks(ctx, strategy.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load strategy stocks: %w", err)
		}
		// Convert []*models.StrategyStock to []models.StrategyStock
		strategy.Stocks = make([]models.StrategyStock, len(stocks))
		for i, stock := range stocks {
			strategy.Stocks[i] = *stock
		}

		strategies = append(strategies, &strategy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategies: %w", err)
	}

	return strategies, nil
}

// Delete deletes a strategy
func (r *strategyRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM strategies WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete strategy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{Resource: "strategy"}
	}

	return nil
}

// AddStockToStrategy adds a stock to a strategy
func (r *strategyRepository) AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error {
	query := `
		INSERT INTO strategy_stocks (strategy_id, stock_id, eligible, created_at)
		VALUES ($1, $2, true, NOW())
		ON CONFLICT (strategy_id, stock_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, strategyID, stockID)
	if err != nil {
		return fmt.Errorf("failed to add stock to strategy: %w", err)
	}

	return nil
}

// RemoveStockFromStrategy removes a stock from a strategy
func (r *strategyRepository) RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error {
	query := `DELETE FROM strategy_stocks WHERE strategy_id = $1 AND stock_id = $2`

	result, err := r.db.ExecContext(ctx, query, strategyID, stockID)
	if err != nil {
		return fmt.Errorf("failed to remove stock from strategy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{Resource: "strategy stock"}
	}

	return nil
}

// UpdateStockEligibility updates the eligibility of a stock within a strategy
func (r *strategyRepository) UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool) error {
	query := `
		UPDATE strategy_stocks 
		SET eligible = $3
		WHERE strategy_id = $1 AND stock_id = $2`

	result, err := r.db.ExecContext(ctx, query, strategyID, stockID, eligible)
	if err != nil {
		return fmt.Errorf("failed to update stock eligibility: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{Resource: "strategy stock"}
	}

	return nil
}

// GetStrategyStocks retrieves all stocks associated with a strategy
func (r *strategyRepository) GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error) {
	query := `
		SELECT ss.strategy_id, ss.stock_id, ss.eligible, ss.created_at,
		       s.id, s.ticker, s.name, s.sector, s.exchange, s.created_at, s.updated_at
		FROM strategy_stocks ss
		JOIN stocks s ON ss.stock_id = s.id
		WHERE ss.strategy_id = $1
		ORDER BY s.ticker`

	rows, err := r.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query strategy stocks: %w", err)
	}
	defer rows.Close()

	var strategyStocks []*models.StrategyStock
	for rows.Next() {
		var ss models.StrategyStock
		var stock models.Stock

		err := rows.Scan(
			&ss.StrategyID,
			&ss.StockID,
			&ss.Eligible,
			&ss.CreatedAt,
			&stock.ID,
			&stock.Ticker,
			&stock.Name,
			&stock.Sector,
			&stock.Exchange,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan strategy stock: %w", err)
		}

		ss.Stock = &stock
		strategyStocks = append(strategyStocks, &ss)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating strategy stocks: %w", err)
	}

	return strategyStocks, nil
}