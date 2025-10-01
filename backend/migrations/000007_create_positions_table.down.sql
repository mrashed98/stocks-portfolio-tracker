-- Drop positions table and related objects
DROP TRIGGER IF EXISTS update_positions_updated_at ON positions;
DROP INDEX IF EXISTS idx_positions_strategy_contrib;
DROP INDEX IF EXISTS idx_positions_stock_id;
DROP INDEX IF EXISTS idx_positions_portfolio_id;
DROP TABLE IF EXISTS positions;