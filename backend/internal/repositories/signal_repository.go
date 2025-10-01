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

// SignalRepository defines the interface for signal data operations
type SignalRepository interface {
	Create(ctx context.Context, signal *models.Signal) (*models.Signal, error)
	Update(ctx context.Context, stockID uuid.UUID, signal models.SignalType) (*models.Signal, error)
	GetCurrentSignal(ctx context.Context, stockID uuid.UUID) (*models.Signal, error)
	GetLatestSignals(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*models.Signal, error)
	GetSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error)
	Delete(ctx context.Context, stockID uuid.UUID, date time.Time) error
}

// signalRepository implements the SignalRepository interface
type signalRepository struct {
	db *sql.DB
}

// NewSignalRepository creates a new signal repository instance
func NewSignalRepository(db *sql.DB) SignalRepository {
	return &signalRepository{db: db}
}

// Create creates a new signal in the database
func (r *signalRepository) Create(ctx context.Context, signal *models.Signal) (*models.Signal, error) {
	query := `
		INSERT INTO signals (stock_id, signal, date, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (stock_id, date) 
		DO UPDATE SET signal = EXCLUDED.signal, created_at = EXCLUDED.created_at
		RETURNING stock_id, signal, date, created_at`

	row := r.db.QueryRowContext(ctx, query,
		signal.StockID,
		signal.Signal,
		signal.Date,
		signal.CreatedAt,
	)

	var created models.Signal
	err := row.Scan(
		&created.StockID,
		&created.Signal,
		&created.Date,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signal: %w", err)
	}

	return &created, nil
}

// Update updates the current signal for a stock (creates a new signal for today)
func (r *signalRepository) Update(ctx context.Context, stockID uuid.UUID, signalType models.SignalType) (*models.Signal, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	signal := &models.Signal{
		StockID:   stockID,
		Signal:    signalType,
		Date:      today,
		CreatedAt: now,
	}

	return r.Create(ctx, signal)
}

// GetCurrentSignal retrieves the most recent signal for a stock
func (r *signalRepository) GetCurrentSignal(ctx context.Context, stockID uuid.UUID) (*models.Signal, error) {
	query := `
		SELECT stock_id, signal, date, created_at
		FROM signals
		WHERE stock_id = $1
		ORDER BY date DESC, created_at DESC
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, stockID)

	var signal models.Signal
	err := row.Scan(
		&signal.StockID,
		&signal.Signal,
		&signal.Date,
		&signal.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{Resource: "signal"}
		}
		return nil, fmt.Errorf("failed to get current signal: %w", err)
	}

	return &signal, nil
}

// GetSignalHistory retrieves signal history for a stock within a date range
func (r *signalRepository) GetSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error) {
	query := `
		SELECT stock_id, signal, date, created_at
		FROM signals
		WHERE stock_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, stockID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query signal history: %w", err)
	}
	defer rows.Close()

	var signals []*models.Signal
	for rows.Next() {
		var signal models.Signal
		err := rows.Scan(
			&signal.StockID,
			&signal.Signal,
			&signal.Date,
			&signal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signal: %w", err)
		}

		signals = append(signals, &signal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating signals: %w", err)
	}

	return signals, nil
}

// Delete deletes a specific signal
func (r *signalRepository) Delete(ctx context.Context, stockID uuid.UUID, date time.Time) error {
	query := `DELETE FROM signals WHERE stock_id = $1 AND date = $2`

	result, err := r.db.ExecContext(ctx, query, stockID, date)
	if err != nil {
		return fmt.Errorf("failed to delete signal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{Resource: "signal"}
	}

	return nil
}

// GetLatestSignals retrieves the most recent signals for multiple stocks
func (r *signalRepository) GetLatestSignals(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*models.Signal, error) {
	if len(stockIDs) == 0 {
		return make(map[uuid.UUID]*models.Signal), nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(stockIDs))
	args := make([]interface{}, len(stockIDs))
	for i, id := range stockIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (stock_id) stock_id, signal, date, created_at
		FROM signals
		WHERE stock_id IN (%s)
		ORDER BY stock_id, date DESC, created_at DESC`, 
		strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest signals: %w", err)
	}
	defer rows.Close()

	signals := make(map[uuid.UUID]*models.Signal)
	for rows.Next() {
		var signal models.Signal
		err := rows.Scan(
			&signal.StockID,
			&signal.Signal,
			&signal.Date,
			&signal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signal: %w", err)
		}

		signals[signal.StockID] = &signal
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating signals: %w", err)
	}

	return signals, nil
}