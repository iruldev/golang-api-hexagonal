package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Story 5.1: Database Pool Validation Tests
// =============================================================================

func TestLoad_DBPool_InvalidMaxConns(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_POOL_MAX_CONNS", "0")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DB_POOL_MAX_CONNS")
}

func TestLoad_DBPool_InvalidMinConns(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_POOL_MIN_CONNS", "-1")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DB_POOL_MIN_CONNS")
}

func TestLoad_DBPool_MinGreaterThanMax(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_POOL_MAX_CONNS", "10")
	t.Setenv("DB_POOL_MIN_CONNS", "20")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DB_POOL_MIN_CONNS")
	assert.Contains(t, err.Error(), "less than or equal to DB_POOL_MAX_CONNS")
}

func TestLoad_DBPool_InvalidMaxLifetime(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_POOL_MAX_LIFETIME", "0s")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DB_POOL_MAX_LIFETIME")
}

func TestLoad_DBPool_ValidConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_POOL_MAX_CONNS", "50")
	t.Setenv("DB_POOL_MIN_CONNS", "50") // Equal should be valid

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, int32(50), cfg.DBPoolMaxConns)
	assert.Equal(t, int32(50), cfg.DBPoolMinConns)
}
