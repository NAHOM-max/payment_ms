package domain

import (
	"context"
	"errors"
	"time"
)

// ErrNonRetryable wraps errors that must skip retries and go straight to DLQ.
var ErrNonRetryable = errors.New("non-retryable error")

type DLQEvent struct {
	OriginalEvent DeliveryConfirmedEvent `json:"original_event"`
	Error         string                 `json:"error"`
	Service       string                 `json:"service"`
	RetryCount    int                    `json:"retry_count"`
	FailedAt      time.Time              `json:"failed_at"`
}

type DLQProducer interface {
	Send(ctx context.Context, key string, event DLQEvent) error
	Close() error
}
