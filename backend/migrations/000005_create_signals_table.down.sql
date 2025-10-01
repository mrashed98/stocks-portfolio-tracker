-- Drop signals table and related objects
DROP INDEX IF EXISTS idx_signals_signal;
DROP INDEX IF EXISTS idx_signals_date;
DROP INDEX IF EXISTS idx_signals_stock_date;
DROP TABLE IF EXISTS signals;