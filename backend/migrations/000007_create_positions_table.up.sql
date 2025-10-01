-- Create positions table for portfolio positions
CREATE TABLE positions (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    stock_id UUID NOT NULL REFERENCES stocks(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    entry_price DECIMAL(10,4) NOT NULL CHECK (entry_price > 0),
    allocation_value DECIMAL(15,2) NOT NULL CHECK (allocation_value > 0),
    strategy_contrib JSONB, -- JSON object mapping strategy_id to contribution amount
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (portfolio_id, stock_id)
);

-- Create indexes for performance
CREATE INDEX idx_positions_portfolio_id ON positions(portfolio_id);
CREATE INDEX idx_positions_stock_id ON positions(stock_id);
CREATE INDEX idx_positions_strategy_contrib ON positions USING GIN (strategy_contrib);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();