#!/bin/bash

# Deployment verification script
set -e

ENVIRONMENT=${1:-production}
BASE_URL=${2:-http://localhost}

echo "🔍 Verifying deployment for $ENVIRONMENT environment..."

# Function to check HTTP endpoint
check_endpoint() {
    local url=$1
    local expected_status=${2:-200}
    local description=$3
    
    echo "Checking $description: $url"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" || echo "000")
    
    if [ "$response" = "$expected_status" ]; then
        echo "✅ $description: OK ($response)"
        return 0
    else
        echo "❌ $description: FAILED ($response)"
        return 1
    fi
}

# Function to check JSON endpoint
check_json_endpoint() {
    local url=$1
    local expected_field=$2
    local description=$3
    
    echo "Checking $description: $url"
    
    response=$(curl -s "$url" || echo "{}")
    
    if echo "$response" | jq -e ".$expected_field" > /dev/null 2>&1; then
        echo "✅ $description: OK"
        return 0
    else
        echo "❌ $description: FAILED - Missing field '$expected_field'"
        echo "Response: $response"
        return 1
    fi
}

# Function to check database connectivity
check_database() {
    echo "Checking database connectivity..."
    
    if docker-compose -f docker-compose.prod.yml exec -T backend ./main health > /dev/null 2>&1; then
        echo "✅ Database connectivity: OK"
        return 0
    else
        echo "❌ Database connectivity: FAILED"
        return 1
    fi
}

# Function to check Redis connectivity
check_redis() {
    echo "Checking Redis connectivity..."
    
    if docker-compose -f docker-compose.prod.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
        echo "✅ Redis connectivity: OK"
        return 0
    else
        echo "❌ Redis connectivity: FAILED"
        return 1
    fi
}

# Function to check container health
check_container_health() {
    echo "Checking container health..."
    
    unhealthy_containers=$(docker-compose -f docker-compose.prod.yml ps --filter "health=unhealthy" -q)
    
    if [ -z "$unhealthy_containers" ]; then
        echo "✅ All containers healthy"
        return 0
    else
        echo "❌ Unhealthy containers found:"
        docker-compose -f docker-compose.prod.yml ps --filter "health=unhealthy"
        return 1
    fi
}

# Function to check resource usage
check_resources() {
    echo "Checking resource usage..."
    
    # Check disk space
    disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 80 ]; then
        echo "✅ Disk usage: ${disk_usage}% (OK)"
    else
        echo "⚠️ Disk usage: ${disk_usage}% (WARNING: High usage)"
    fi
    
    # Check memory usage
    memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ "$memory_usage" -lt 80 ]; then
        echo "✅ Memory usage: ${memory_usage}% (OK)"
    else
        echo "⚠️ Memory usage: ${memory_usage}% (WARNING: High usage)"
    fi
}

# Function to run performance tests
run_performance_tests() {
    echo "Running basic performance tests..."
    
    # Test API response time
    start_time=$(date +%s%N)
    check_endpoint "$BASE_URL/api/health" 200 "API response time test" > /dev/null
    end_time=$(date +%s%N)
    
    response_time=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
    
    if [ "$response_time" -lt 1000 ]; then
        echo "✅ API response time: ${response_time}ms (OK)"
    else
        echo "⚠️ API response time: ${response_time}ms (WARNING: Slow response)"
    fi
}

# Main verification process
echo "Starting deployment verification..."
echo "Environment: $ENVIRONMENT"
echo "Base URL: $BASE_URL"
echo "Timestamp: $(date)"
echo "----------------------------------------"

failed_checks=0

# Basic health checks
check_endpoint "$BASE_URL/health" 200 "Application health" || ((failed_checks++))
check_endpoint "$BASE_URL/api/health" 200 "API health" || ((failed_checks++))

# API endpoint checks
check_json_endpoint "$BASE_URL/api/v1/" "message" "API root endpoint" || ((failed_checks++))

# Infrastructure checks (only if running locally with docker-compose)
if [ -f "docker-compose.prod.yml" ]; then
    check_container_health || ((failed_checks++))
    check_database || ((failed_checks++))
    check_redis || ((failed_checks++))
    check_resources || ((failed_checks++))
fi

# Performance checks
run_performance_tests || ((failed_checks++))

# Security checks
echo "Checking security headers..."
security_headers=$(curl -s -I "$BASE_URL" | grep -E "(X-Frame-Options|X-XSS-Protection|X-Content-Type-Options)" | wc -l)
if [ "$security_headers" -ge 2 ]; then
    echo "✅ Security headers: OK"
else
    echo "⚠️ Security headers: Missing some security headers"
    ((failed_checks++))
fi

# SSL check (if HTTPS)
if [[ "$BASE_URL" == https* ]]; then
    echo "Checking SSL certificate..."
    if curl -s --head "$BASE_URL" > /dev/null 2>&1; then
        echo "✅ SSL certificate: OK"
    else
        echo "❌ SSL certificate: FAILED"
        ((failed_checks++))
    fi
fi

# Summary
echo "----------------------------------------"
echo "Verification Summary:"
echo "Environment: $ENVIRONMENT"
echo "Failed checks: $failed_checks"

if [ "$failed_checks" -eq 0 ]; then
    echo "🎉 All checks passed! Deployment is healthy."
    exit 0
else
    echo "❌ $failed_checks check(s) failed. Please investigate."
    exit 1
fi