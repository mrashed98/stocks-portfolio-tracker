# Production Deployment Guide

This guide covers deploying the Portfolio Web App to production using Docker containers.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 4GB RAM and 20GB disk space
- SSL certificates (for HTTPS deployment)

## Quick Start

1. **Clone and prepare environment:**
   ```bash
   git clone <repository-url>
   cd portfolio-manager
   cp .env.prod .env.prod.local
   ```

2. **Configure environment variables:**
   Edit `.env.prod.local` with your production values:
   ```bash
   # Database Configuration
   DB_NAME=portfolio_db_prod
   DB_USER=portfolio_user_prod
   DB_PASSWORD=your_secure_db_password_here
   
   # Redis Configuration
   REDIS_PASSWORD=your_secure_redis_password_here
   
   # JWT Configuration
   JWT_SECRET=your_very_secure_jwt_secret_key_here_minimum_32_characters
   
   # Market Data API
   MARKET_DATA_API_KEY=your_market_data_api_key_here
   
   # Application URLs
   API_URL=https://your-domain.com/api
   ```

3. **Deploy to production:**
   ```bash
   ./scripts/deploy-prod.sh
   ```

## Architecture Overview

The production deployment consists of:

- **Frontend**: React app served by Nginx
- **Backend**: Go API server
- **Database**: PostgreSQL with persistent storage
- **Cache**: Redis for market data caching
- **Reverse Proxy**: Nginx for load balancing and SSL termination
- **Monitoring**: Prometheus + Grafana (optional)

## Security Features

### Network Security
- Internal networks for backend services
- Only necessary ports exposed
- Rate limiting on API endpoints
- CORS protection

### Container Security
- Non-root users in all containers
- Multi-stage builds to minimize attack surface
- Health checks for all services
- Resource limits to prevent DoS

### Data Security
- Environment-based secret management
- Encrypted database connections
- JWT token authentication
- Password hashing with bcrypt

## SSL/HTTPS Configuration

### Using Let's Encrypt (Recommended)

1. **Install Certbot:**
   ```bash
   sudo apt-get install certbot
   ```

2. **Generate certificates:**
   ```bash
   sudo certbot certonly --standalone -d your-domain.com
   ```

3. **Copy certificates:**
   ```bash
   sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/cert.pem
   sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/key.pem
   ```

4. **Update nginx configuration:**
   Uncomment the HTTPS server block in `nginx/nginx.conf`

### Using Custom Certificates

1. **Place certificates:**
   ```bash
   cp your-cert.pem nginx/ssl/cert.pem
   cp your-key.pem nginx/ssl/key.pem
   ```

2. **Update nginx configuration:**
   Uncomment and configure the HTTPS server block

## Monitoring and Logging

### Enable Monitoring Stack

```bash
# Start monitoring services
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana
open http://localhost:3001
# Default credentials: admin/admin
```

### Log Management

```bash
# View application logs
docker-compose -f docker-compose.prod.yml logs -f

# View specific service logs
docker-compose -f docker-compose.prod.yml logs -f backend

# Export logs for analysis
docker-compose -f docker-compose.prod.yml logs --since 24h > logs/app-$(date +%Y%m%d).log
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U $DB_USER $DB_NAME > backup-$(date +%Y%m%d).sql

# Restore from backup
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U $DB_USER $DB_NAME < backup-20231201.sql
```

### Volume Backup

```bash
# Backup all volumes
docker run --rm -v portfolio-manager_postgres_data:/data -v $(pwd)/backups:/backup alpine tar czf /backup/postgres-$(date +%Y%m%d).tar.gz -C /data .
docker run --rm -v portfolio-manager_redis_data:/data -v $(pwd)/backups:/backup alpine tar czf /backup/redis-$(date +%Y%m%d).tar.gz -C /data .
```

## Scaling and Performance

### Horizontal Scaling

```bash
# Scale backend service
docker-compose -f docker-compose.prod.yml up -d --scale backend=3

# Update nginx upstream configuration for load balancing
```

### Performance Tuning

1. **Database Optimization:**
   - Adjust PostgreSQL configuration in `postgresql.conf`
   - Monitor query performance with `pg_stat_statements`
   - Set up connection pooling with PgBouncer

2. **Redis Optimization:**
   - Configure memory limits and eviction policies
   - Enable persistence for critical cached data
   - Monitor memory usage and hit rates

3. **Application Optimization:**
   - Enable Go profiling endpoints
   - Monitor garbage collection metrics
   - Optimize database queries and indexes

## Troubleshooting

### Common Issues

1. **Services not starting:**
   ```bash
   # Check service status
   docker-compose -f docker-compose.prod.yml ps
   
   # Check logs for errors
   docker-compose -f docker-compose.prod.yml logs
   ```

2. **Database connection issues:**
   ```bash
   # Test database connectivity
   docker-compose -f docker-compose.prod.yml exec backend ./main health
   
   # Check database logs
   docker-compose -f docker-compose.prod.yml logs postgres
   ```

3. **Memory issues:**
   ```bash
   # Monitor resource usage
   docker stats
   
   # Check container limits
   docker-compose -f docker-compose.prod.yml config
   ```

### Health Checks

All services include health checks accessible via:

- **Application**: `http://localhost/health`
- **Backend API**: `http://localhost/api/health`
- **Database**: Automatic via Docker health check
- **Redis**: Automatic via Docker health check

### Emergency Procedures

1. **Rollback deployment:**
   ```bash
   # Stop current deployment
   docker-compose -f docker-compose.prod.yml down
   
   # Restore from backup
   # ... restore database and volumes ...
   
   # Deploy previous version
   git checkout previous-tag
   ./scripts/deploy-prod.sh
   ```

2. **Scale down for maintenance:**
   ```bash
   # Graceful shutdown
   docker-compose -f docker-compose.prod.yml stop
   
   # Maintenance mode (serve static page)
   docker run -d -p 80:80 -v $(pwd)/maintenance.html:/usr/share/nginx/html/index.html nginx:alpine
   ```

## Security Checklist

- [ ] Environment variables configured securely
- [ ] SSL certificates installed and configured
- [ ] Database passwords are strong and unique
- [ ] JWT secret is cryptographically secure
- [ ] API rate limiting is enabled
- [ ] Container images are regularly updated
- [ ] Backup procedures are tested
- [ ] Monitoring and alerting are configured
- [ ] Log retention policies are set
- [ ] Network access is restricted to necessary ports

## Maintenance

### Regular Tasks

1. **Weekly:**
   - Review application logs for errors
   - Check disk space and clean old logs
   - Verify backup integrity

2. **Monthly:**
   - Update container images
   - Review security patches
   - Analyze performance metrics
   - Test disaster recovery procedures

3. **Quarterly:**
   - Security audit and penetration testing
   - Capacity planning review
   - Update SSL certificates if needed
   - Review and update documentation

### Update Procedure

```bash
# 1. Backup current state
./scripts/backup.sh

# 2. Pull latest changes
git pull origin main

# 3. Update images
docker-compose -f docker-compose.prod.yml pull

# 4. Deploy updates
./scripts/deploy-prod.sh

# 5. Verify deployment
curl -f http://localhost/health || echo "Deployment failed"
```