package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"portfolio-app/internal/models"
)

// StockRepository defines the interface for stock data operations
type StockRepository interface {
	Create(ctx context.Context, stock *models.Stock) (*models.Stock, error)
	Update(ctx context.Context, stock *models.Stock) (*models.Stock, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Stock, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Stock, error)
	GetByTicker(ctx context.Context, ticker string) (*models.Stock, error)
	GetAll(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetStocksWithCurrentSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error)
}

// stockRepository implements the StockRepository interface
type stockRepository struct {
	db *sql.DB
}

// NewStockRepository creates a new stock repository instance
func NewStockRepository(db *sql.DB) StockRepository {
	return &stockRepository{db: db}
}

// Create creates a new stock in the database
func (r *stockRepository) Create(ctx context.Context, stock *models.Stock) (*models.Stock, error) {
	query := `
		INSERT INTO stocks (id, ticker, name, sector, exchange, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, ticker, name, sector, exchange, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		stock.ID,
		stock.Ticker,
		stock.Name,
		stock.Sector,
		stock.Exchange,
		stock.CreatedAt,
		stock.UpdatedAt,
	)

	var created models.Stock
	err := row.Scan(
		&created.ID,
		&created.Ticker,
		&created.Name,
		&created.Sector,
		&created.Exchange,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock: %w", err)
	}

	return &created, nil
}

// Update updates an existing stock in the database
func (r *stockRepository) Update(ctx context.Context, stock *models.Stock) (*models.Stock, error) {
	query := `
		UPDATE stocks 
		SET name = $2, sector = $3, exchange = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, ticker, name, sector, exchange, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		stock.ID,
		stock.Name,
		stock.Sector,
		stock.Exchange,
		stock.UpdatedAt,
	)

	var updated models.Stock
	err := row.Scan(
		&updated.ID,
		&updated.Ticker,
		&updated.Name,
		&updated.Sector,
		&updated.Exchange,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "stock"}
		}
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return &updated, nil
}

// GetByID retrieves a stock by ID
func (r *stockRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	query := `
		SELECT id, ticker, name, sector, exchange, created_at, updated_at
		FROM stocks
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	var stock models.Stock
	err := row.Scan(
		&stock.ID,
		&stock.Ticker,
		&stock.Name,
		&stock.Sector,
		&stock.Exchange,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "stock"}
		}
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

// GetByIDs retrieves multiple stocks by their IDs
func (r *stockRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Stock, error) {
	if len(ids) == 0 {
		return []*models.Stock{}, nil
	}

	var stocks []*models.Stock
	
	// For simplicity, query each stock individually
	for _, id := range ids {
		query := `
			SELECT id, ticker, name, sector, exchange, created_at, updated_at
			FROM stocks
			WHERE id = $1`

		row := r.db.QueryRowContext(ctx, query, id)

		stock := &models.Stock{}
		err := row.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.Name,
			&stock.Sector,
			&stock.Exchange,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				continue // Skip missing stocks
			}
			return nil, fmt.Errorf("failed to get stock %s: %w", id, err)
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

// GetByTicker retrieves a stock by ticker symbol
func (r *stockRepository) GetByTicker(ctx context.Context, ticker string) (*models.Stock, error) {
	query := `
		SELECT id, ticker, name, sector, exchange, created_at, updated_at
		FROM stocks
		WHERE ticker = $1`

	row := r.db.QueryRowContext(ctx, query, strings.ToUpper(ticker))

	var stock models.Stock
	err := row.Scan(
		&stock.ID,
		&stock.Ticker,
		&stock.Name,
		&stock.Sector,
		&stock.Exchange,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "stock"}
		}
		return nil, fmt.Errorf("failed to get stock by ticker: %w", err)
	}

	return &stock, nil
}

// GetAll retrieves all stocks with optional search and pagination
func (r *stockRepository) GetAll(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT id, ticker, name, sector, exchange, created_at, updated_at
			FROM stocks
			WHERE ticker ILIKE $1 OR name ILIKE $1 OR sector ILIKE $1
			ORDER BY ticker
			LIMIT $2 OFFSET $3`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = `
			SELECT id, ticker, name, sector, exchange, created_at, updated_at
			FROM stocks
			ORDER BY ticker
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.Name,
			&stock.Sector,
			&stock.Exchange,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}

		stocks = append(stocks, &stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// Delete deletes a stock
func (r *stockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM stocks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{Resource: "stock"}
	}

	return nil
}

// GetStocksWithCurrentSignals retrieves stocks with their current signals
func (r *stockRepository) GetStocksWithCurrentSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	var query string
	var args []interface{}

	baseQuery := `
		SELECT DISTINCT s.id, s.ticker, s.name, s.sector, s.exchange, s.created_at, s.updated_at,
		       sig.stock_id, sig.signal, sig.date, sig.created_at as signal_created_at
		FROM stocks s
		LEFT JOIN LATERAL (
			SELECT stock_id, signal, date, created_at
			FROM signals
			WHERE stock_id = s.id
			ORDER BY date DESC, created_at DESC
			LIMIT 1
		) sig ON true`

	if search != "" {
		query = baseQuery + `
			WHERE s.ticker ILIKE $1 OR s.name ILIKE $1 OR s.sector ILIKE $1
			ORDER BY s.ticker
			LIMIT $2 OFFSET $3`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = baseQuery + `
			ORDER BY s.ticker
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks with signals: %w", err)
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		var stock models.Stock
		var signal models.Signal
		var signalStockID *uuid.UUID
		var signalType *models.SignalType
		var signalDate *time.Time
		var signalCreatedAt *time.Time

		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.Name,
			&stock.Sector,
			&stock.Exchange,
			&stock.CreatedAt,
			&stock.UpdatedAt,
			&signalStockID,
			&signalType,
			&signalDate,
			&signalCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock with signal: %w", err)
		}

		// If signal exists, populate it
		if signalStockID != nil {
			signal.StockID = *signalStockID
			signal.Signal = *signalType
			signal.Date = *signalDate
			signal.CreatedAt = *signalCreatedAt
			stock.CurrentSignal = &signal
		}

		stocks = append(stocks, &stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks with signals: %w", err)
	}

	return stocks, nil
}