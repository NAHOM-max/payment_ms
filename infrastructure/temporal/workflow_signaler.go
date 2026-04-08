package temporal

import (
	"context"

	"go.temporal.io/sdk/client"

	"payment_ms/domain"
)

const signalName = "payment_update"

type SignalPayload struct {
	PaymentID string
	Status    domain.PaymentStatus
}

type WorkflowSignaler struct {
	client client.Client
}

func NewWorkflowSignaler(c client.Client) *WorkflowSignaler {
	return &WorkflowSignaler{client: c}
}

func (s *WorkflowSignaler) Signal(ctx context.Context, workflowID string, status domain.PaymentStatus, paymentID string) error {
	return s.client.SignalWorkflow(ctx, workflowID, "", signalName, SignalPayload{
		PaymentID: paymentID,
		Status:    status,
	})
}
