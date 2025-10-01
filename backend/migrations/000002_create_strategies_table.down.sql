-- Drop strategies table and related objects
DROP TRIGGER IF EXISTS update_strategies_updated_at ON strategies;
DROP INDEX IF EXISTS idx_strategies_user_name;
DROP INDEX IF EXISTS idx_strategies_user_id;
DROP TABLE IF EXISTS strategies;