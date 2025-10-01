package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"portfolio-app/config"
)

func TestNewConnection(t *testing.T) {
	// Skip if no test database is available
	if os.Getenv("TEST_DB_URL") == "" {
		t.Skip("TEST_DB_URL not set, skipping database tests")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "portfolio_test",
			User:     "test_user",
			Password: "test_pass",
			SSLMode:  "disable",
		},
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Test connection is working
	err = db.Health()
	assert.NoError(t, err)

	// Test connection pool settings
	stats := db.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
}

func TestDatabaseHealth(t *testing.T) {
	if os.Getenv("TEST_DB_URL") == "" {
		t.Skip("TEST_DB_URL not set, skipping database tests")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "portfolio_test",
			User:     "test_user",
			Password: "test_pass",
			SSLMode:  "disable",
		},
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Test health check
	err = db.Health()
	assert.NoError(t, err)

	// Test health check after closing
	db.Close()
	err = db.Health()
	assert.Error(t, err)
}

func TestConnectionPooling(t *testing.T) {
	if os.Getenv("TEST_DB_URL") == "" {
		t.Skip("TEST_DB_URL not set, skipping database tests")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "portfolio_test",
			User:     "test_user",
			Password: "test_pass",
			SSLMode:  "disable",
		},
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Test multiple concurrent connections
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			var result int
			err := db.QueryRow("SELECT 1").Scan(&result)
			assert.NoError(t, err)
			assert.Equal(t, 1, result)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent connections")
		}
	}
}

// Integration test for database connection with invalid credentials
func TestConnectionFailure(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "nonexistent_db",
			User:     "invalid_user",
			Password: "invalid_pass",
			SSLMode:  "disable",
		},
	}

	_, err := NewConnection(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")
}