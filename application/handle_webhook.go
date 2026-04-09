package application

import (
	"context"
	"errors"

	"payment_ms/domain"
)

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// validTransitions defines allowed next states from a given current status.
// Terminal states (SUCCESSFUL, FAILED, CANCELED) are intentionally absent —
// no transition out of them is permitted.
var validTransitions = map[domain.PaymentStatus]map[domain.PaymentStatus]bool{
	domain.StatusInitiated: {
		domain.StatusSuccessful: true,
		domain.StatusFailed:     true,
		domain.StatusCanceled:   true,
	},
}

type HandleWebhookInput struct {
	PaymentID string
	Status    domain.PaymentStatus
}

type HandleWebhookOutput struct {
	WorkflowID string
	Status     domain.PaymentStatus
}

type HandleWebhookUseCase struct {
	repo     domain.PaymentRepository
	signaler WorkflowSignaler
}

func NewHandleWebhookUseCase(repo domain.PaymentRepository, signaler WorkflowSignaler) *HandleWebhookUseCase {
	return &HandleWebhookUseCase{repo: repo, signaler: signaler}
}

func (uc *HandleWebhookUseCase) Execute(ctx context.Context, input HandleWebhookInput) (*HandleWebhookOutput, error) {
	payment, err := uc.repo.GetByID(ctx, input.PaymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Idempotency: duplicate delivery of the same terminal status is a no-op.
	if payment.Status == input.Status {
		return &HandleWebhookOutput{
			WorkflowID: payment.WorkflowID,
			Status:     payment.Status,
		}, nil
	}

	allowed, ok := validTransitions[payment.Status]
	if !ok || !allowed[input.Status] {
		return nil, ErrInvalidStatusTransition
	}

	// Conditional update: only applies if the DB row still has the expected
	// current status, guarding against concurrent webhook deliveries.
	updated, err := uc.repo.UpdateStatusConditional(ctx, input.PaymentID, payment.Status, input.Status)
	if err != nil {
		return nil, err
	}
	if !updated {
		// Another request already transitioned this payment; treat as idempotent.
		return &HandleWebhookOutput{
			WorkflowID: payment.WorkflowID,
			Status:     input.Status,
		}, nil
	}

	if err := uc.signaler.Signal(ctx, payment.WorkflowID, input.Status, input.PaymentID); err != nil {
		return nil, err
	}

	return &HandleWebhookOutput{
		WorkflowID: payment.WorkflowID,
		Status:     input.Status,
	}, nil
}
