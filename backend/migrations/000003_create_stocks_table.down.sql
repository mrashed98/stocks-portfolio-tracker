-- Drop stocks table and related objects
DROP TRIGGER IF EXISTS update_stocks_updated_at ON stocks;
DROP INDEX IF EXISTS idx_stocks_sector;
DROP INDEX IF EXISTS idx_stocks_ticker;
DROP TABLE IF EXISTS stocks;