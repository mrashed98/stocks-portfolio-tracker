-- Create strategy-stock relationships table
CREATE TABLE strategy_stocks (
    strategy_id UUID NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
    stock_id UUID NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    eligible BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (strategy_id, stock_id)
);

-- Create indexes for performance
CREATE INDEX idx_strategy_stocks_strategy_id ON strategy_stocks(strategy_id);
CREATE INDEX idx_strategy_stocks_stock_id ON strategy_stocks(stock_id);
CREATE INDEX idx_strategy_stocks_eligible ON strategy_stocks(strategy_id, eligible);