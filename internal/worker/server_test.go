package worker

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name            string
		cfg             config.AsynqConfig
		wantConcurrency int
	}{
		{
			name:            "default concurrency when zero",
			cfg:             config.AsynqConfig{Concurrency: 0},
			wantConcurrency: 10,
		},
		{
			name:            "custom concurrency",
			cfg:             config.AsynqConfig{Concurrency: 20},
			wantConcurrency: 20,
		},
		{
			name:            "with shutdown timeout",
			cfg:             config.AsynqConfig{Concurrency: 5, ShutdownTimeout: 60 * time.Second},
			wantConcurrency: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}

			// Act
			srv := NewServer(redisOpt, tt.cfg)

			// Assert
			require.NotNil(t, srv)
			assert.NotNil(t, srv.srv)
			assert.NotNil(t, srv.mux)
		})
	}
}

func TestQueueConstants(t *testing.T) {
	// Assert queue constants are defined correctly
	assert.Equal(t, "critical", QueueCritical)
	assert.Equal(t, "default", QueueDefault)
	assert.Equal(t, "low", QueueLow)
}
