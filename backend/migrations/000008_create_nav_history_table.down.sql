-- Drop nav_history table and related objects
DROP INDEX IF EXISTS idx_nav_history_timestamp;
DROP INDEX IF EXISTS idx_nav_history_portfolio_timestamp;
DROP TABLE IF EXISTS nav_history;