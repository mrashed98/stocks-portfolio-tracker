# Implementation Plan

- [x] 1. Set up project structure and development environment
  - Create Docker Compose configuration for PostgreSQL, Redis, and development services
  - Initialize Go module with Fiber framework and required dependencies
  - Set up React TypeScript project with Tailwind CSS and shadcn/ui
  - Create environment configuration management for database connections and API keys
  - _Requirements: 6.1, 6.2_

- [x] 2. Implement database schema and migrations
  - Create PostgreSQL migration files for all core tables (users, strategies, stocks, etc.)
  - Implement database connection utilities with connection pooling
  - Create database seeding scripts for development data
  - Write unit tests for database schema validation
  - _Requirements: 6.2, 7.4_

- [x] 3. Create core data models and validation
- [x] 3.1 Implement Go structs for all domain entities
  - Define User, Strategy, Stock, Portfolio, Position, and NAVHistory structs
  - Add JSON tags and validation rules using validator package
  - Create model conversion functions between database and API representations
  - _Requirements: 1.1, 2.1, 3.1, 4.1_

- [x] 3.2 Implement TypeScript interfaces for frontend models
  - Create TypeScript interfaces matching Go structs for type safety
  - Define API request and response types for all endpoints
  - Implement client-side validation schemas using Zod
  - _Requirements: 1.1, 2.1, 3.1, 4.1_

- [x] 4. Build Strategy Management system
- [x] 4.1 Implement Strategy Service backend
  - Create StrategyService interface and implementation with CRUD operations
  - Implement strategy weight validation (percentage mode total ≤ 100%)
  - Add database repository layer for strategy persistence
  - Write unit tests for strategy business logic
  - _Requirements: 1.1, 1.2, 1.5_

- [x] 4.2 Create Strategy API endpoints
  - Implement POST /strategies endpoint for strategy creation
  - Create PUT /strategies/:id/stocks endpoint for stock assignment
  - Add PUT /strategies/:id/weight endpoint for weight updates
  - Implement proper error handling and validation responses
  - _Requirements: 7.1, 7.4_

- [x] 4.3 Build Strategy Designer frontend component
  - Create React component for strategy creation and editing
  - Implement real-time weight validation with user feedback
  - Add stock selection interface with eligibility toggles
  - Integrate with backend API using React Query for state management
  - _Requirements: 1.1, 1.3, 1.5_

- [x] 5. Implement Stock Management system
- [x] 5.1 Create Stock Service and Signal Management
  - Implement StockService with ticker validation and basic info fetching
  - Create signal management system for Buy/Hold status tracking
  - Add stock eligibility management across multiple strategies
  - Write unit tests for stock assignment and signal logic
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 5.2 Build Stock API endpoints
  - Create GET /stocks endpoint with search and filtering
  - Implement POST /stocks endpoint for adding new stocks
  - Add PUT /stocks/:id/signal endpoint for signal updates
  - Create stock eligibility management endpoints
  - _Requirements: 7.1, 2.1_

- [x] 6. Develop Allocation Engine
- [x] 6.1 Implement core allocation calculation algorithm
  - Create AllocationEngine with strategy weight distribution logic
  - Implement stock-level allocation within strategies for "Buy" signal stocks
  - Add constraint application (min/max allocations per stock)
  - Create allocation normalization to match total investment amount
  - Add allocation rebalancing to re-allocate investments after thier value changes overtime
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 6.2 Add constraint validation and error handling
  - Implement constraint validator for maximum allocation percentages
  - Add minimum allocation amount validation
  - Create detailed error messages for constraint violations with suggestions
  - Write comprehensive unit tests for edge cases and constraint scenarios
  - _Requirements: 3.4, 7.4_

- [x] 6.3 Build allocation preview functionality
  - Create allocation preview generation with real-time price fetching
  - Implement quantity calculation with floor logic and unallocated cash tracking
  - Add dynamic recalculation when stocks are removed from strategies
  - Create allocation preview API endpoint with caching
  - _Requirements: 3.5, 3.6, 2.4_

- [x] 7. Implement Market Data Service
- [x] 7.1 Create Market Data Service with external API integration with tradingview
  - Search the web for already created packages fetch data from tradingview
  - Implement MarketDataService interface with quote fetching
  - Add Redis caching layer for market data with TTL management
  - Create circuit breaker pattern for external API resilience
  - Implement batch quote fetching for portfolio updates
  - _Requirements: 5.2, 8.3, 7.4_

- [x] 7.2 Build Market Data API endpoints
  - Create GET /quotes/:ticker endpoint with caching
  - Implement GET /ohlcv/:ticker endpoint for historical data
  - Add batch quote endpoint for portfolio position updates
  - Create TradingView DataFeed compatible endpoints (/symbols, /history, /search)
  - _Requirements: 7.3, 5.1, 5.3_

- [x] 8. Develop Portfolio Management system
- [x] 8.1 Implement Portfolio Service backend
  - Create PortfolioService with allocation preview and portfolio creation
  - Implement portfolio snapshot creation with entry prices and positions
  - Add NAV calculation logic with P&L and drawdown metrics
  - Create portfolio history tracking and retrieval
  - Add ability to re-weight the investment to follow strategy allocations
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 8.2 Build Portfolio API endpoints
  - Create POST /portfolios/preview endpoint for allocation calculations
  - Implement POST /portfolios endpoint for portfolio creation with snapshots
  - Add GET /portfolios/:id endpoint for portfolio details
  - Create GET /portfolios/:id/history endpoint for performance tracking
  - _Requirements: 7.2, 4.1, 4.4_

- [x] 8.3 Create Portfolio Builder frontend component
  - Build React component for portfolio creation with strategy selection
  - Implement real-time allocation preview with dynamic updates
  - Add stock removal/restoration functionality with instant recalculation
  - Create portfolio creation flow with confirmation and snapshot display
  - Add ability to re-weight the investment to follow strategy allocations
  - _Requirements: 3.5, 2.4, 4.1_

- [x] 9. Build Portfolio Tracking and Performance system
- [x] 9.1 Implement background NAV update scheduler
  - Create background job system using Go routines and cron scheduling
  - Implement portfolio NAV update logic with current market prices
  - Add P&L calculation and maximum drawdown tracking
  - Create error handling and retry logic for failed updates
  - _Requirements: 8.1, 8.2, 8.5_

- [x] 9.2 Create Portfolio Performance frontend components
  - Build portfolio dashboard with NAV history charts
  - Implement performance metrics display (returns, drawdown, P&L)
  - Add portfolio position details with current values and performance
  - Create data staleness indicators for market data age
  - _Requirements: 4.4, 8.3_

- [x] 10. Integrate TradingView Charts
- [x] 10.1 Implement TradingView DataFeed backend
  - Create DataFeed API endpoints compatible with TradingView Charting Library
  - Implement symbol search, resolution, and historical data endpoints
  - Add real-time data streaming capabilities for live charts
  - Create error handling for chart data unavailability
  - _Requirements: 5.3, 5.4_

- [x] 10.2 Build TradingView Chart frontend component
  - Create React wrapper component for TradingView Charting Library
  - Implement chart configuration with custom DataFeed integration
  - Add chart theming and responsive design
  - Create fallback UI for chart loading failures
  - _Requirements: 5.1, 5.4_

- [x] 11. Implement Authentication and User Management
- [x] 11.1 Create User Service and authentication
  - Implement user registration and login functionality
  - Add JWT token generation and validation middleware
  - Create user session management with Redis
  - Implement password hashing and security best practices
  - _Requirements: 7.5, 6.3_

- [x] 11.2 Add authentication to all API endpoints
  - Implement JWT middleware for protected routes
  - Add user context extraction for multi-tenant data isolation
  - Create rate limiting middleware for API protection
  - Add CORS configuration for frontend integration
  - _Requirements: 7.5, 6.3_

- [x] 12. Build main application UI and routing
- [x] 12.1 Create main application layout and navigation
  - Implement React Router setup with protected routes
  - Create main navigation component with user authentication state
  - Add responsive layout with sidebar navigation
  - Implement loading states and error boundaries
  - _Requirements: 6.4_

- [x] 12.2 Integrate all components into complete application
  - Connect Strategy Designer to main application flow
  - Integrate Portfolio Builder with strategy selection
  - Add Portfolio Dashboard with performance tracking
  - Create stock detail pages with TradingView charts
  - _Requirements: 1.1, 3.1, 4.1, 5.1_

- [x] 13. Add comprehensive testing suite
- [x] 13.1 Write unit tests for all backend services
  - Create unit tests for AllocationEngine with various scenarios
  - Add tests for StrategyService, PortfolioService, and MarketDataService
  - Implement database integration tests using testcontainers
  - Create mock implementations for external API testing
  - _Requirements: 6.4, 7.4_

- [x] 13.2 Implement frontend component testing
  - Write React Testing Library tests for all major components
  - Add integration tests for API interactions
  - Create end-to-end tests for complete user workflows
  - Implement visual regression testing for UI consistency
  - _Requirements: 6.4_

- [x] 14. Containerization and deployment setup
- [x] 14.1 Create production Docker configuration
  - Build multi-stage Dockerfiles for frontend and backend
  - Create production Docker Compose with proper networking
  - Add health checks and monitoring for all services
  - Implement proper secret management and environment configuration
  - _Requirements: 6.1, 6.4_

- [x] 14.2 Set up CI/CD pipeline
  - Create GitHub Actions workflow for automated testing
  - Add Docker image building and registry pushing
  - Implement automated deployment to staging environment
  - Create database migration automation in deployment pipeline
  - _Requirements: 6.4_