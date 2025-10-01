# Portfolio Management Web Application

A containerized web application for creating and managing investment portfolios through a strategy-based approach.

## Architecture

- **Frontend**: React 18 + TypeScript, Tailwind CSS, shadcn/ui components
- **Backend**: Go with Fiber framework
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Containerization**: Docker with Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Development Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd portfolio-app
```

2. Copy environment files:
```bash
cp .env.example .env
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

3. Update environment variables in the `.env` files as needed.

4. Start the development environment:
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database on port 5432
- Redis cache on port 6379
- Go backend API on port 8080
- React frontend on port 3000

5. Access the application:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### Development Workflow

The development setup includes hot reloading for both frontend and backend:

- **Frontend**: Uses Vite dev server with hot module replacement
- **Backend**: Uses Air for automatic Go binary rebuilding on file changes

### Project Structure

```
portfolio-app/
├── backend/                 # Go backend application
│   ├── config/             # Configuration management
│   ├── main.go             # Application entry point
│   ├── go.mod              # Go module dependencies
│   └── Dockerfile.dev      # Development Docker image
├── frontend/               # React frontend application
│   ├── src/                # Source code
│   │   ├── components/     # React components
│   │   ├── pages/          # Page components
│   │   ├── lib/            # Utilities and configuration
│   │   └── hooks/          # Custom React hooks
│   ├── package.json        # Node.js dependencies
│   └── Dockerfile.dev      # Development Docker image
├── docker-compose.yml      # Development services
└── README.md              # This file
```

### Environment Variables

Key environment variables:

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`: Database connection
- `REDIS_HOST`, `REDIS_PORT`: Redis connection
- `JWT_SECRET`: JWT token signing secret
- `MARKET_DATA_API_KEY`: External market data API key
- `REACT_APP_API_URL`: Frontend API endpoint

### Next Steps

After the development environment is running, you can:

1. Implement database migrations (Task 2)
2. Create data models and validation (Task 3)
3. Build the Strategy Management system (Task 4)
4. Continue with the remaining implementation tasks

## Development Commands

### Backend

```bash
# Run backend locally (requires Go 1.21+)
cd backend
go run main.go

# Run tests
go test ./...

# Install dependencies
go mod tidy
```

### Frontend

```bash
# Run frontend locally (requires Node.js 18+)
cd frontend
npm install
npm run dev

# Build for production
npm run build

# Run linting
npm run lint
```

### Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild images
docker-compose build
```