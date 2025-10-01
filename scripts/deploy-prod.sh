#!/bin/bash

# Production deployment script for Portfolio Web App
set -e

echo "🚀 Starting production deployment..."

# Check if required environment file exists
if [ ! -f ".env.prod.local" ]; then
    echo "❌ Error: .env.prod.local file not found"
    echo "Please copy .env.prod to .env.prod.local and configure with production values"
    exit 1
fi

# Load production environment variables
export $(cat .env.prod.local | grep -v '^#' | xargs)

# Validate required environment variables
required_vars=("DB_NAME" "DB_USER" "DB_PASSWORD" "REDIS_PASSWORD" "JWT_SECRET" "MARKET_DATA_API_KEY")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Error: Required environment variable $var is not set"
        exit 1
    fi
done

echo "✅ Environment variables validated"

# Create necessary directories
mkdir -p nginx/ssl
mkdir -p logs

# Build and deploy with Docker Compose
echo "🔨 Building Docker images..."
docker-compose -f docker-compose.prod.yml build --no-cache

echo "🔄 Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

echo "🗄️ Creating Docker volumes..."
docker-compose -f docker-compose.prod.yml up --no-start

echo "🚀 Starting production services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
timeout=300
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if docker-compose -f docker-compose.prod.yml ps | grep -q "healthy"; then
        echo "✅ Services are healthy"
        break
    fi
    sleep 5
    elapsed=$((elapsed + 5))
    echo "Waiting... ($elapsed/$timeout seconds)"
done

if [ $elapsed -ge $timeout ]; then
    echo "❌ Timeout waiting for services to be healthy"
    docker-compose -f docker-compose.prod.yml logs
    exit 1
fi

# Run database migrations
echo "🗄️ Running database migrations..."
docker-compose -f docker-compose.prod.yml exec -T backend ./main migrate

# Display deployment status
echo "📊 Deployment Status:"
docker-compose -f docker-compose.prod.yml ps

echo "🎉 Production deployment completed successfully!"
echo "📱 Application is available at: http://localhost"
echo "🔍 API health check: http://localhost/api/health"

# Show logs for monitoring
echo "📝 Showing recent logs (press Ctrl+C to exit):"
docker-compose -f docker-compose.prod.yml logs -f --tail=50