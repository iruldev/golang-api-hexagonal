package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPGXPoolConfig_ApplyConfig(t *testing.T) {
	databaseURL := "postgres://user:pass@localhost:5432/testdb"
	poolCfg := PoolConfig{
		MaxConns:        50,
		MinConns:        10,
		MaxConnLifetime: 30 * time.Minute,
	}

	config, err := getPGXPoolConfig(databaseURL, poolCfg)

	require.NoError(t, err)
	assert.Equal(t, int32(50), config.MaxConns)
	assert.Equal(t, int32(10), config.MinConns)
	assert.Equal(t, 30*time.Minute, config.MaxConnLifetime)
}

func TestGetPGXPoolConfig_DefaultsOrZeroWait(t *testing.T) {
	// If we pass 0, it should supposedly respect the defaults or whatever ParseConfig does (defaults, usually)
	// getPGXPoolConfig DOES NOT overwrite if 0.
	databaseURL := "postgres://user:pass@localhost:5432/testdb"
	poolCfg := PoolConfig{
		MaxConns:        0,
		MinConns:        0,
		MaxConnLifetime: 0,
	}

	config, err := getPGXPoolConfig(databaseURL, poolCfg)

	require.NoError(t, err)
	// Verify defaults from pgxpool (check docs or assume defaults)
	// pgxpool default MaxConns is usually max(4, runtime.NumCPU())
	assert.Greater(t, config.MaxConns, int32(0))
	// We just want to ensure our 0 didn't overwrite with 0 if that's invalid,
	// but actually pgxpool might allow 0 for some things?
	// Our logic: "if poolCfg.MaxConns > 0 { set it }"
	// So it should remain whatever ParseConfig returned.
}
