-- Drop portfolios table and related objects
DROP TRIGGER IF EXISTS update_portfolios_updated_at ON portfolios;
DROP INDEX IF EXISTS idx_portfolios_user_name;
DROP INDEX IF EXISTS idx_portfolios_user_id;
DROP TABLE IF EXISTS portfolios;