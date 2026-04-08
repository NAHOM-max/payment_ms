package application

import (
	"context"

	"payment_ms/domain"
)

// WorkflowSignaler is the outbound port for signaling a Temporal workflow.
// Implemented in infrastructure, consumed here via dependency inversion.
type WorkflowSignaler interface {
	Signal(ctx context.Context, workflowID string, status domain.PaymentStatus, paymentID string) error
}
