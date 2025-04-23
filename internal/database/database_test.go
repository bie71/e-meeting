package database

import (
	"context"
	"testing"
	"time"

	"e_metting/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestContainer(t *testing.T) (*config.Config, func()) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	require.NoError(t, err)

	// Get connection details
	hostIP, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := &config.Config{
		DBHost:               hostIP,
		DBPort:               mappedPort.Int(),
		DBUser:               "postgres",
		DBPassword:           "postgres",
		DBName:               "test_db",
		DBMaxOpenConnections: 25,
		DBMaxIdleConnections: 5,
	}

	cleanup := func() {
		require.NoError(t, container.Terminate(ctx))
	}

	return cfg, cleanup
}

func TestNew(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	db := New(cfg)
	require.NotNil(t, db)

	// Test Health
	stats := db.Health()
	assert.Equal(t, "up", stats["status"])

	// Test DB
	sqlDB := db.DB()
	require.NotNil(t, sqlDB)

	// Test GormDB
	gormDB := db.GormDB()
	require.NotNil(t, gormDB)

	// Test Close
	err := db.Close()
	assert.NoError(t, err)
}

func TestHealth(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	db := New(cfg)
	stats := db.Health()

	assert.Contains(t, stats, "status")
	assert.Contains(t, stats, "message")
	assert.Contains(t, stats, "open_connections")
	assert.Contains(t, stats, "in_use")
	assert.Contains(t, stats, "idle")
}

func TestClose(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	db := New(cfg)
	err := db.Close()
	assert.NoError(t, err)

	// Test that we can't use the connection after closing
	err = db.DB().Ping()
	assert.Error(t, err)
}

func TestConnectionPool(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	db := New(cfg)
	sqlDB := db.DB()

	// Test connection pool stats
	stats := sqlDB.Stats()
	assert.Equal(t, 25, stats.MaxOpenConnections)
	assert.Equal(t, 5, stats.Idle)
}

func TestConcurrentAccess(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	// Create multiple instances
	db1 := New(cfg)
	db2 := New(cfg)

	// They should be the same instance
	assert.Equal(t, db1, db2)

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			stats := db1.Health()
			assert.Equal(t, "up", stats["status"])
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestContextTimeout(t *testing.T) {
	cfg, cleanup := setupTestContainer(t)
	defer cleanup()

	db := New(cfg)
	sqlDB := db.DB()

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should fail due to timeout
	err := sqlDB.PingContext(ctx)
	assert.Error(t, err)
}
