-- Create signals table for stock signals
CREATE TABLE signals (
    stock_id UUID NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    signal VARCHAR(10) NOT NULL CHECK (signal IN ('Buy', 'Hold')),
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (stock_id, date)
);

-- Create indexes for performance
CREATE INDEX idx_signals_stock_date ON signals(stock_id, date DESC);
CREATE INDEX idx_signals_date ON signals(date DESC);
CREATE INDEX idx_signals_signal ON signals(signal);