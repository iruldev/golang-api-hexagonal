package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client for enqueueing tasks.
type Client struct {
	cli *asynq.Client
}

// NewClient creates a new Asynq client with the given Redis options.
func NewClient(redisOpt asynq.RedisClientOpt) *Client {
	return &Client{
		cli: asynq.NewClient(redisOpt),
	}
}

// Enqueue adds a task to the specified queue with options.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.cli.Enqueue(task, opts...)
}

// EnqueueCritical adds a task to the critical queue.
func (c *Client) EnqueueCritical(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error) {
	return c.cli.Enqueue(task, asynq.Queue(QueueCritical))
}

// EnqueueDefault adds a task to the default queue.
func (c *Client) EnqueueDefault(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error) {
	return c.cli.Enqueue(task, asynq.Queue(QueueDefault))
}

// EnqueueLow adds a task to the low priority queue.
func (c *Client) EnqueueLow(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error) {
	return c.cli.Enqueue(task, asynq.Queue(QueueLow))
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.cli.Close()
}
