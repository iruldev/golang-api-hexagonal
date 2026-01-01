package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockPingable is a mock implementation of the pingable interface.
type mockPingable struct {
	mock.Mock
}

func (m *mockPingable) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestDatabaseHealthChecker_CheckHealth(t *testing.T) {
	t.Run("returns healthy when ping succeeds", func(t *testing.T) {
		mockPool := new(mockPingable)
		mockPool.On("Ping", mock.Anything).Return(nil)

		checker := NewDatabaseHealthChecker(mockPool)
		status, latency, err := checker.CheckHealth(context.Background())

		assert.Equal(t, "healthy", status)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, latency, time.Duration(0))
		mockPool.AssertExpectations(t)
	})

	t.Run("returns unhealthy when ping fails", func(t *testing.T) {
		mockPool := new(mockPingable)
		expectedErr := errors.New("connection failed")
		mockPool.On("Ping", mock.Anything).Return(expectedErr)

		checker := NewDatabaseHealthChecker(mockPool)
		status, latency, err := checker.CheckHealth(context.Background())

		assert.Equal(t, "unhealthy", status)
		assert.Equal(t, expectedErr, err)
		assert.GreaterOrEqual(t, latency, time.Duration(0))
		mockPool.AssertExpectations(t)
	})
}

func TestDatabaseHealthChecker_Name(t *testing.T) {
	mockPool := new(mockPingable)
	checker := NewDatabaseHealthChecker(mockPool)
	assert.Equal(t, "database", checker.Name())
}

func TestNewDatabaseCheck(t *testing.T) {
	t.Run("returns nil when ping succeeds", func(t *testing.T) {
		mockPool := new(mockPingable)
		mockPool.On("Ping", mock.Anything).Return(nil)

		check := NewDatabaseCheck(mockPool, 100*time.Millisecond)
		err := check()

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("returns error when ping fails", func(t *testing.T) {
		mockPool := new(mockPingable)
		expectedErr := errors.New("connection failed")
		mockPool.On("Ping", mock.Anything).Return(expectedErr)

		check := NewDatabaseCheck(mockPool, 100*time.Millisecond)
		err := check()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockPool.AssertExpectations(t)
	})
}
