package resilience

import (
	"errors"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultResilienceConfig(t *testing.T) {
	cfg := DefaultResilienceConfig()

	// Circuit Breaker defaults
	assert.Equal(t, DefaultCBMaxRequests, cfg.CircuitBreaker.MaxRequests)
	assert.Equal(t, DefaultCBInterval, cfg.CircuitBreaker.Interval)
	assert.Equal(t, DefaultCBTimeout, cfg.CircuitBreaker.Timeout)
	assert.Equal(t, DefaultCBFailureThreshold, cfg.CircuitBreaker.FailureThreshold)

	// Retry defaults
	assert.Equal(t, DefaultRetryMaxAttempts, cfg.Retry.MaxAttempts)
	assert.Equal(t, DefaultRetryInitialDelay, cfg.Retry.InitialDelay)
	assert.Equal(t, DefaultRetryMaxDelay, cfg.Retry.MaxDelay)
	assert.Equal(t, DefaultRetryMultiplier, cfg.Retry.Multiplier)

	// Timeout defaults
	assert.Equal(t, DefaultTimeoutDefault, cfg.Timeout.Default)
	assert.Equal(t, DefaultTimeoutDatabase, cfg.Timeout.Database)
	assert.Equal(t, DefaultTimeoutExternalAPI, cfg.Timeout.ExternalAPI)

	// Bulkhead defaults
	assert.Equal(t, DefaultBulkheadMaxConcurrent, cfg.Bulkhead.MaxConcurrent)
	assert.Equal(t, DefaultBulkheadMaxWaiting, cfg.Bulkhead.MaxWaiting)

	// Shutdown defaults
	assert.Equal(t, DefaultShutdownDrainPeriod, cfg.Shutdown.DrainPeriod)
	assert.Equal(t, DefaultShutdownGracePeriod, cfg.Shutdown.GracePeriod)
}

func TestResilienceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ResilienceConfig
		wantErr string
	}{
		{
			name:    "valid defaults",
			config:  DefaultResilienceConfig(),
			wantErr: "",
		},
		{
			name: "valid custom config",
			config: ResilienceConfig{
				CircuitBreaker: CircuitBreakerConfig{
					MaxRequests:      5,
					Interval:         20 * time.Second,
					Timeout:          60 * time.Second,
					FailureThreshold: 10,
				},
				Retry: RetryConfig{
					MaxAttempts:  5,
					InitialDelay: 200 * time.Millisecond,
					MaxDelay:     10 * time.Second,
					Multiplier:   3.0,
				},
				Timeout: TimeoutConfig{
					Default:     60 * time.Second,
					Database:    10 * time.Second,
					ExternalAPI: 20 * time.Second,
				},
				Bulkhead: BulkheadConfig{
					MaxConcurrent: 20,
					MaxWaiting:    200,
				},
				Shutdown: ShutdownConfig{
					DrainPeriod: 45 * time.Second,
					GracePeriod: 10 * time.Second,
				},
			},
			wantErr: "",
		},
		// Circuit Breaker validation
		{
			name: "invalid cb max_requests zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.MaxRequests = 0
				return cfg
			}(),
			wantErr: "max_requests must be greater than 0",
		},
		{
			name: "invalid cb max_requests negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.MaxRequests = -1
				return cfg
			}(),
			wantErr: "max_requests must be greater than 0",
		},
		{
			name: "invalid cb interval zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.Interval = 0
				return cfg
			}(),
			wantErr: "interval must be greater than 0",
		},
		{
			name: "invalid cb timeout zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.Timeout = 0
				return cfg
			}(),
			wantErr: "timeout must be greater than 0",
		},
		{
			name: "invalid cb failure_threshold zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.FailureThreshold = 0
				return cfg
			}(),
			wantErr: "failure_threshold must be greater than 0",
		},
		// Retry validation
		{
			name: "invalid retry max_attempts zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.MaxAttempts = 0
				return cfg
			}(),
			wantErr: "max_attempts must be greater than 0",
		},
		{
			name: "invalid retry initial_delay zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.InitialDelay = 0
				return cfg
			}(),
			wantErr: "initial_delay must be greater than 0",
		},
		{
			name: "invalid retry max_delay zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.MaxDelay = 0
				return cfg
			}(),
			wantErr: "max_delay must be greater than 0",
		},
		{
			name: "invalid retry max_delay less than initial_delay",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.InitialDelay = 5 * time.Second
				cfg.Retry.MaxDelay = 1 * time.Second
				return cfg
			}(),
			wantErr: "max_delay must be greater than or equal to initial_delay",
		},
		{
			name: "invalid retry multiplier less than 1",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.Multiplier = 0.5
				return cfg
			}(),
			wantErr: "multiplier must be greater than or equal to 1.0",
		},
		// Timeout validation
		{
			name: "invalid timeout default zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.Default = 0
				return cfg
			}(),
			wantErr: "default timeout must be greater than 0",
		},
		{
			name: "invalid timeout database zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.Database = 0
				return cfg
			}(),
			wantErr: "database timeout must be greater than 0",
		},
		{
			name: "invalid timeout external_api zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.ExternalAPI = 0
				return cfg
			}(),
			wantErr: "external_api timeout must be greater than 0",
		},
		// Bulkhead validation
		{
			name: "invalid bulkhead max_concurrent zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Bulkhead.MaxConcurrent = 0
				return cfg
			}(),
			wantErr: "max_concurrent must be greater than 0",
		},
		{
			name: "invalid bulkhead max_waiting negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Bulkhead.MaxWaiting = -1
				return cfg
			}(),
			wantErr: "max_waiting must be non-negative",
		},
		// Edge case: bulkhead max_waiting zero is valid
		{
			name: "valid bulkhead max_waiting zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Bulkhead.MaxWaiting = 0
				return cfg
			}(),
			wantErr: "",
		},
		// Shutdown validation (Story 1.6)
		{
			name: "invalid shutdown drain_period zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Shutdown.DrainPeriod = 0
				return cfg
			}(),
			wantErr: "drain_period must be greater than 0",
		},
		{
			name: "invalid shutdown drain_period negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Shutdown.DrainPeriod = -1 * time.Second
				return cfg
			}(),
			wantErr: "drain_period must be greater than 0",
		},
		{
			name: "invalid shutdown grace_period negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Shutdown.GracePeriod = -1 * time.Second
				return cfg
			}(),
			wantErr: "grace_period must be non-negative",
		},
		// Edge case: shutdown grace_period zero is valid (no extra grace time)
		{
			name: "valid shutdown grace_period zero",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Shutdown.GracePeriod = 0
				return cfg
			}(),
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResilienceError(t *testing.T) {
	t.Run("Error returns formatted message", func(t *testing.T) {
		err := ErrCircuitOpen
		assert.Equal(t, "RES-001: circuit breaker is open", err.Error())
	})

	t.Run("Error with underlying error includes it", func(t *testing.T) {
		underlyingErr := errors.New("connection refused")
		err := NewCircuitOpenError(underlyingErr)
		assert.Contains(t, err.Error(), "RES-001")
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlyingErr := errors.New("connection refused")
		err := NewTimeoutExceededError(underlyingErr)
		var resErr *ResilienceError
		require.True(t, errors.As(err, &resErr))
		assert.Equal(t, underlyingErr, resErr.Unwrap())
	})

	t.Run("errors.Is matches by code", func(t *testing.T) {
		err := NewCircuitOpenError(nil)
		assert.True(t, errors.Is(err, ErrCircuitOpen))
		assert.False(t, errors.Is(err, ErrBulkheadFull))
		assert.False(t, errors.Is(err, ErrTimeoutExceeded))
		assert.False(t, errors.Is(err, ErrMaxRetriesExceeded))
	})

	t.Run("errors.As extracts ResilienceError", func(t *testing.T) {
		err := NewMaxRetriesExceededError(errors.New("all attempts failed"))
		var resErr *ResilienceError
		require.True(t, errors.As(err, &resErr))
		assert.Equal(t, ErrCodeMaxRetriesExceeded, resErr.Code)
	})

	t.Run("nil ResilienceError returns <nil>", func(t *testing.T) {
		var err *ResilienceError
		assert.Equal(t, "<nil>", err.Error())
	})
}

func TestNewResilienceConfig(t *testing.T) {
	cfg := &config.Config{
		// Circuit Breaker
		CBMaxRequests:      5,
		CBInterval:         20 * time.Second,
		CBTimeout:          60 * time.Second,
		CBFailureThreshold: 10,
		// Retry
		RetryMaxAttempts:  5,
		RetryInitialDelay: 200 * time.Millisecond,
		RetryMaxDelay:     10 * time.Second,
		RetryMultiplier:   3.0,
		// Timeout
		TimeoutDefault:     60 * time.Second,
		TimeoutDatabase:    10 * time.Second,
		TimeoutExternalAPI: 20 * time.Second,
		// Bulkhead
		BulkheadMaxConcurrent: 20,
		BulkheadMaxWaiting:    200,
		// Shutdown
		ShutdownDrainPeriod: 45 * time.Second,
		ShutdownGracePeriod: 10 * time.Second,
	}

	resCfg := NewResilienceConfig(cfg)

	// Verify Circuit Breaker mapping
	assert.Equal(t, cfg.CBMaxRequests, resCfg.CircuitBreaker.MaxRequests)
	assert.Equal(t, cfg.CBInterval, resCfg.CircuitBreaker.Interval)
	assert.Equal(t, cfg.CBTimeout, resCfg.CircuitBreaker.Timeout)
	assert.Equal(t, cfg.CBFailureThreshold, resCfg.CircuitBreaker.FailureThreshold)

	// Verify Retry mapping
	assert.Equal(t, cfg.RetryMaxAttempts, resCfg.Retry.MaxAttempts)
	assert.Equal(t, cfg.RetryInitialDelay, resCfg.Retry.InitialDelay)
	assert.Equal(t, cfg.RetryMaxDelay, resCfg.Retry.MaxDelay)
	assert.Equal(t, cfg.RetryMultiplier, resCfg.Retry.Multiplier)

	// Verify Timeout mapping
	assert.Equal(t, cfg.TimeoutDefault, resCfg.Timeout.Default)
	assert.Equal(t, cfg.TimeoutDatabase, resCfg.Timeout.Database)
	assert.Equal(t, cfg.TimeoutExternalAPI, resCfg.Timeout.ExternalAPI)

	// Verify Bulkhead mapping
	assert.Equal(t, cfg.BulkheadMaxConcurrent, resCfg.Bulkhead.MaxConcurrent)
	assert.Equal(t, cfg.BulkheadMaxWaiting, resCfg.Bulkhead.MaxWaiting)

	// Verify Shutdown mapping
	assert.Equal(t, cfg.ShutdownDrainPeriod, resCfg.Shutdown.DrainPeriod)
	assert.Equal(t, cfg.ShutdownGracePeriod, resCfg.Shutdown.GracePeriod)
}

func TestResilienceConfig_Validate_NegativeDurations(t *testing.T) {
	tests := []struct {
		name    string
		config  ResilienceConfig
		wantErr string
	}{
		{
			name: "invalid cb interval negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.Interval = -1 * time.Second
				return cfg
			}(),
			wantErr: "interval must be greater than 0",
		},
		{
			name: "invalid cb timeout negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.CircuitBreaker.Timeout = -1 * time.Second
				return cfg
			}(),
			wantErr: "timeout must be greater than 0",
		},
		{
			name: "invalid retry initial_delay negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.InitialDelay = -1 * time.Millisecond
				return cfg
			}(),
			wantErr: "initial_delay must be greater than 0",
		},
		{
			name: "invalid retry max_delay negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Retry.MaxDelay = -1 * time.Second
				return cfg
			}(),
			wantErr: "max_delay must be greater than 0",
		},
		{
			name: "invalid timeout default negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.Default = -1 * time.Second
				return cfg
			}(),
			wantErr: "default timeout must be greater than 0",
		},
		{
			name: "invalid timeout database negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.Database = -1 * time.Second
				return cfg
			}(),
			wantErr: "database timeout must be greater than 0",
		},
		{
			name: "invalid timeout external_api negative",
			config: func() ResilienceConfig {
				cfg := DefaultResilienceConfig()
				cfg.Timeout.ExternalAPI = -1 * time.Second
				return cfg
			}(),
			wantErr: "external_api timeout must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestResilienceError_Is_NonResilienceError(t *testing.T) {
	t.Run("errors.Is returns false for non-ResilienceError target", func(t *testing.T) {
		err := NewCircuitOpenError(nil)
		nonResErr := errors.New("some other error")
		assert.False(t, errors.Is(err, nonResErr))
	})

	t.Run("errors.Is returns false when comparing different codes", func(t *testing.T) {
		err1 := NewCircuitOpenError(nil)
		err2 := NewBulkheadFullError(nil)
		assert.False(t, errors.Is(err1, err2))
	})
}

func TestNewResilienceErrors(t *testing.T) {
	tests := []struct {
		name         string
		createErr    func(error) error
		expectedCode string
	}{
		{
			name:         "NewCircuitOpenError",
			createErr:    NewCircuitOpenError,
			expectedCode: ErrCodeCircuitOpen,
		},
		{
			name:         "NewBulkheadFullError",
			createErr:    NewBulkheadFullError,
			expectedCode: ErrCodeBulkheadFull,
		},
		{
			name:         "NewTimeoutExceededError",
			createErr:    NewTimeoutExceededError,
			expectedCode: ErrCodeTimeoutExceeded,
		},
		{
			name:         "NewMaxRetriesExceededError",
			createErr:    NewMaxRetriesExceededError,
			expectedCode: ErrCodeMaxRetriesExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			underlyingErr := errors.New("underlying error")
			err := tt.createErr(underlyingErr)

			var resErr *ResilienceError
			require.True(t, errors.As(err, &resErr))
			assert.Equal(t, tt.expectedCode, resErr.Code)
			assert.Equal(t, underlyingErr, resErr.Err)
		})
	}
}
