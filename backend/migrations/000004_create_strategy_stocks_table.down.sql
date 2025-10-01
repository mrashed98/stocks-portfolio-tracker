-- Drop strategy_stocks table and related objects
DROP INDEX IF EXISTS idx_strategy_stocks_eligible;
DROP INDEX IF EXISTS idx_strategy_stocks_stock_id;
DROP INDEX IF EXISTS idx_strategy_stocks_strategy_id;
DROP TABLE IF EXISTS strategy_stocks;