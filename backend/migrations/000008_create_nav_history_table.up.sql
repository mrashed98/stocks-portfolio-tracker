-- Create NAV history table for performance tracking
CREATE TABLE nav_history (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    nav DECIMAL(15,2) NOT NULL CHECK (nav >= 0),
    pnl DECIMAL(15,2) NOT NULL,
    drawdown DECIMAL(8,4), -- Percentage drawdown from peak
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (portfolio_id, timestamp)
);

-- Create indexes for performance
CREATE INDEX idx_nav_history_portfolio_timestamp ON nav_history(portfolio_id, timestamp DESC);
CREATE INDEX idx_nav_history_timestamp ON nav_history(timestamp DESC);