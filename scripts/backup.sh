#!/bin/bash

# Backup script for Portfolio Web App
set -e

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

echo "🗄️ Starting backup process..."

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Load environment variables
if [ -f ".env.prod.local" ]; then
    export $(cat .env.prod.local | grep -v '^#' | xargs)
fi

# Database backup
echo "📊 Backing up PostgreSQL database..."
docker-compose -f docker-compose.prod.yml exec -T postgres pg_dump -U "${DB_USER:-portfolio_user_prod}" "${DB_NAME:-portfolio_db_prod}" > "$BACKUP_DIR/postgres_$DATE.sql"

# Redis backup
echo "🔄 Backing up Redis data..."
docker-compose -f docker-compose.prod.yml exec -T redis redis-cli --rdb /data/dump.rdb BGSAVE
sleep 5  # Wait for background save to complete
docker run --rm -v portfolio-manager_redis_data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine cp /data/dump.rdb "/backup/redis_$DATE.rdb"

# Volume backups
echo "💾 Backing up Docker volumes..."
docker run --rm -v portfolio-manager_postgres_data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf "/backup/postgres_volume_$DATE.tar.gz" -C /data .
docker run --rm -v portfolio-manager_redis_data:/data -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf "/backup/redis_volume_$DATE.tar.gz" -C /data .

# Configuration backup
echo "⚙️ Backing up configuration files..."
tar czf "$BACKUP_DIR/config_$DATE.tar.gz" \
    .env.prod.local \
    docker-compose.prod.yml \
    nginx/ \
    monitoring/ \
    --exclude='nginx/ssl/*.key' 2>/dev/null || true

# Create backup manifest
echo "📋 Creating backup manifest..."
cat > "$BACKUP_DIR/manifest_$DATE.txt" << EOF
Portfolio Web App Backup
Date: $(date)
Backup ID: $DATE

Files included:
- postgres_$DATE.sql (Database dump)
- redis_$DATE.rdb (Redis data)
- postgres_volume_$DATE.tar.gz (PostgreSQL volume)
- redis_volume_$DATE.tar.gz (Redis volume)
- config_$DATE.tar.gz (Configuration files)

Restore instructions:
1. Stop all services: docker-compose -f docker-compose.prod.yml down
2. Restore database: docker-compose -f docker-compose.prod.yml exec -T postgres psql -U \$DB_USER \$DB_NAME < postgres_$DATE.sql
3. Restore volumes: docker run --rm -v portfolio-manager_postgres_data:/data -v \$(pwd)/backups:/backup alpine tar xzf /backup/postgres_volume_$DATE.tar.gz -C /data
4. Restart services: docker-compose -f docker-compose.prod.yml up -d
EOF

# Cleanup old backups (keep last 7 days)
echo "🧹 Cleaning up old backups..."
find "$BACKUP_DIR" -name "*.sql" -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "*.rdb" -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -name "manifest_*.txt" -mtime +7 -delete 2>/dev/null || true

# Calculate backup size
BACKUP_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)

echo "✅ Backup completed successfully!"
echo "📁 Backup location: $BACKUP_DIR"
echo "📏 Total backup size: $BACKUP_SIZE"
echo "🆔 Backup ID: $DATE"

# Optional: Upload to cloud storage
if [ "$CLOUD_BACKUP_ENABLED" = "true" ]; then
    echo "☁️ Uploading to cloud storage..."
    # Add your cloud upload commands here
    # Example for AWS S3:
    # aws s3 sync "$BACKUP_DIR" "s3://your-backup-bucket/portfolio-app/$DATE/"
fi