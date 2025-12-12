package patterns

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScheduledJob_Fields(t *testing.T) {
	// Arrange
	task := asynq.NewTask("test:task", []byte(`{"key":"value"}`))

	// Act
	job := ScheduledJob{
		Cronspec:    "0 0 * * *",
		Task:        task,
		Description: "Test job description",
		Opts:        []asynq.Option{asynq.MaxRetry(5)},
	}

	// Assert
	assert.Equal(t, "0 0 * * *", job.Cronspec)
	assert.Equal(t, "test:task", job.Task.Type())
	assert.Equal(t, "Test job description", job.Description)
	assert.Len(t, job.Opts, 1)
}

func TestScheduledJob_EmptyOptional(t *testing.T) {
	// Arrange - job with minimal required fields
	task := asynq.NewTask("test:minimal", nil)

	// Act
	job := ScheduledJob{
		Cronspec: "* * * * *",
		Task:     task,
	}

	// Assert
	assert.Equal(t, "* * * * *", job.Cronspec)
	assert.Empty(t, job.Description)
	assert.Nil(t, job.Opts)
}

func TestRegisterScheduledJobs_EmptyList(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	var jobs []ScheduledJob

	// Act - should handle empty list gracefully
	// Note: We can't test with a real scheduler without Redis,
	// so we just verify the function handles empty input
	entryIDs, err := RegisterScheduledJobs(nil, jobs, logger)

	// Assert - empty list returns empty without error
	require.NoError(t, err)
	assert.Empty(t, entryIDs)
}

func TestCronExpressions_CommonPatterns(t *testing.T) {
	// This test documents and validates common cron patterns
	// These are pattern validation tests, not scheduler integration tests
	tests := []struct {
		name     string
		cronspec string
		meaning  string
	}{
		{
			name:     "every minute",
			cronspec: "* * * * *",
			meaning:  "runs every minute",
		},
		{
			name:     "every hour",
			cronspec: "0 * * * *",
			meaning:  "runs at minute 0 of every hour",
		},
		{
			name:     "daily at midnight",
			cronspec: "0 0 * * *",
			meaning:  "runs at 00:00 every day",
		},
		{
			name:     "every monday at 9am",
			cronspec: "0 9 * * 1",
			meaning:  "runs at 09:00 every Monday",
		},
		{
			name:     "first of month",
			cronspec: "0 0 1 * *",
			meaning:  "runs at midnight on the 1st of every month",
		},
		{
			name:     "every 5 minutes",
			cronspec: "*/5 * * * *",
			meaning:  "runs every 5 minutes",
		},
		{
			name:     "weekdays at 6pm",
			cronspec: "0 18 * * 1-5",
			meaning:  "runs at 18:00 Monday through Friday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a job with this cronspec
			job := ScheduledJob{
				Cronspec:    tt.cronspec,
				Task:        asynq.NewTask("test:cron", nil),
				Description: tt.meaning,
			}

			// Verify job is properly constructed
			assert.Equal(t, tt.cronspec, job.Cronspec)
			assert.Equal(t, tt.meaning, job.Description)
		})
	}
}

func TestScheduledJob_WithOptions(t *testing.T) {
	// Arrange
	task := asynq.NewTask("test:with-opts", nil)

	// Act
	job := ScheduledJob{
		Cronspec:    "0 0 * * *",
		Task:        task,
		Description: "Job with multiple options",
		Opts: []asynq.Option{
			asynq.MaxRetry(5),
			asynq.Queue("critical"),
		},
	}

	// Assert
	assert.Len(t, job.Opts, 2)
}

func TestValidateCronspec_ValidExpressions(t *testing.T) {
	tests := []struct {
		name     string
		cronspec string
	}{
		{"every minute", "* * * * *"},
		{"every hour", "0 * * * *"},
		{"daily at midnight", "0 0 * * *"},
		{"monday at 9am", "0 9 * * 1"},
		{"first of month", "0 0 1 * *"},
		{"every 5 minutes", "*/5 * * * *"},
		{"weekdays at 6pm", "0 18 * * 1-5"},
		{"complex expression", "15,45 9-17 * * 1-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCronspec(tt.cronspec)
			assert.NoError(t, err)
		})
	}
}

func TestValidateCronspec_InvalidExpressions(t *testing.T) {
	tests := []struct {
		name     string
		cronspec string
	}{
		{"empty string", ""},
		{"gibberish", "invalid"},
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * *"},
		{"invalid minute", "60 * * * *"},
		{"invalid hour", "0 25 * * *"},
		{"invalid day of month", "0 0 32 * *"},
		{"invalid month", "0 0 * 13 *"},
		{"invalid day of week", "0 0 * * 8"},
		{"negative values", "-1 * * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCronspec(tt.cronspec)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid cron expression")
		})
	}
}
