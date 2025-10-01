# Requirements Document

## Introduction

The Integrated Portfolio Web App is a containerized web application that enables users to create and manage investment portfolios through a strategy-based approach. Users can define modular investment strategies, assign stocks to multiple strategies, and dynamically calculate portfolio allocations with real-time constraints. The application provides portfolio tracking, performance monitoring, and integrated market data visualization through TradingView charts.

## Requirements

### Requirement 1

**User Story:** As an investor, I want to create and manage investment strategies, so that I can organize my investment approach in a modular way.

#### Acceptance Criteria

1. WHEN a user creates a new strategy THEN the system SHALL allow them to specify a name and weight mode (percentage or fixed budget)
2. WHEN a user assigns a weight to a strategy THEN the system SHALL validate that percentage weights don't exceed 100% across all strategies
3. WHEN a user adds stocks to a strategy THEN the system SHALL allow them to toggle stock eligibility within that strategy
4. IF a stock belongs to multiple strategies THEN the system SHALL track separate eligibility flags for each strategy
5. WHEN a user modifies strategy weights THEN the system SHALL recalculate all dependent portfolio allocations in real-time

### Requirement 2

**User Story:** As an investor, I want to manage stocks with signals and multi-strategy assignments, so that I can implement complex investment approaches.

#### Acceptance Criteria

1. WHEN a user adds a stock THEN the system SHALL require a valid ticker symbol and automatically fetch basic stock information
2. WHEN a stock has a signal THEN the system SHALL only include "Buy" signal stocks in allocation calculations
3. WHEN a user removes a stock from a strategy THEN the system SHALL automatically redistribute allocations among remaining eligible stocks
4. IF a stock is assigned to multiple strategies THEN the system SHALL calculate combined allocation weights across all strategies
5. WHEN stock eligibility changes THEN the system SHALL update portfolio allocations within 1 second

### Requirement 3

**User Story:** As an investor, I want to build portfolios with dynamic allocation calculations, so that I can see how my investment will be distributed before committing funds.

#### Acceptance Criteria

1. WHEN a user inputs a total investment amount THEN the system SHALL calculate allocations based on strategy weights and stock eligibility
2. WHEN calculating allocations THEN the system SHALL apply maximum allocation constraints per stock (percentage of total portfolio)
3. WHEN calculating allocations THEN the system SHALL ensure minimum allocation amounts per stock are met
4. IF allocation constraints cannot be satisfied THEN the system SHALL display clear error messages with suggested adjustments
5. WHEN generating allocation preview THEN the system SHALL show stock ticker, weight, allocation amount, and calculated quantity
6. WHEN fetching live prices THEN the system SHALL calculate quantities as floor(allocation / price) and track unallocated cash

### Requirement 4

**User Story:** As an investor, I want to create portfolio snapshots and track performance over time, so that I can monitor my investment returns and make informed decisions.

#### Acceptance Criteria

1. WHEN a user creates a portfolio THEN the system SHALL save a snapshot with entry prices, quantities, and allocation details
2. WHEN a portfolio is created THEN the system SHALL record the creation timestamp and initial NAV
3. WHEN market data updates THEN the system SHALL recalculate portfolio NAV, P&L, and drawdown metrics
4. WHEN displaying portfolio performance THEN the system SHALL show historical NAV, percentage returns, and maximum drawdown
5. IF market data is unavailable THEN the system SHALL use the last known prices and display a staleness indicator

### Requirement 5

**User Story:** As an investor, I want to view interactive stock charts and market data, so that I can analyze stocks before making investment decisions.

#### Acceptance Criteria

1. WHEN a user views a stock THEN the system SHALL display an interactive TradingView chart with price history
2. WHEN displaying market data THEN the system SHALL show current price, daily change, and basic fundamentals
3. WHEN charts load THEN the system SHALL fetch historical OHLCV data through a custom DataFeed implementation
4. IF TradingView integration fails THEN the system SHALL display basic price information and error message
5. WHEN viewing portfolio details THEN the system SHALL provide chart access for all portfolio positions

### Requirement 6

**User Story:** As a system administrator, I want the application to be fully containerized and scalable, so that it can be deployed reliably across different environments.

#### Acceptance Criteria

1. WHEN deploying the application THEN the system SHALL run entirely within Docker containers
2. WHEN the application starts THEN the system SHALL automatically initialize the database schema and required tables
3. WHEN handling concurrent users THEN the system SHALL maintain data consistency through proper transaction management
4. IF any service fails THEN the system SHALL provide graceful degradation and clear error messages
5. WHEN scaling the application THEN the system SHALL support horizontal scaling through stateless service design

### Requirement 7

**User Story:** As a developer, I want comprehensive APIs for all core functionality, so that the system can be extended and integrated with other tools.

#### Acceptance Criteria

1. WHEN creating strategies THEN the API SHALL provide endpoints for CRUD operations with proper validation
2. WHEN managing portfolios THEN the API SHALL provide preview calculations before portfolio creation
3. WHEN fetching market data THEN the API SHALL provide real-time quotes and historical data endpoints
4. IF API requests fail THEN the system SHALL return appropriate HTTP status codes and error messages
5. WHEN accessing APIs THEN the system SHALL implement proper authentication and rate limiting

### Requirement 8

**User Story:** As an investor, I want real-time portfolio updates and background processing, so that my portfolio data stays current without manual intervention.

#### Acceptance Criteria

1. WHEN market hours are active THEN the system SHALL update portfolio NAV at least every 15 minutes
2. WHEN background jobs run THEN the system SHALL update all portfolio positions without affecting user experience
3. IF price data becomes stale THEN the system SHALL indicate data age and attempt to refresh
4. WHEN system load is high THEN the system SHALL prioritize user-facing operations over background updates
5. WHEN errors occur in background processing THEN the system SHALL log errors and retry with exponential backoff