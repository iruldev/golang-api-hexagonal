package worker

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		redisOpt asynq.RedisClientOpt
	}{
		{
			name:     "creates client with valid redis opt",
			redisOpt: asynq.RedisClientOpt{Addr: "localhost:6379"},
		},
		{
			name:     "creates client with password",
			redisOpt: asynq.RedisClientOpt{Addr: "localhost:6379", Password: "secret"},
		},
		{
			name:     "creates client with db selection",
			redisOpt: asynq.RedisClientOpt{Addr: "localhost:6379", DB: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			client := NewClient(tt.redisOpt)

			// Assert
			require.NotNil(t, client)
			assert.NotNil(t, client.cli)

			// Cleanup
			_ = client.Close()
		})
	}
}

func TestClientClose(t *testing.T) {
	// Arrange
	client := NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	require.NotNil(t, client)

	// Act
	err := client.Close()

	// Assert
	assert.NoError(t, err)
}
