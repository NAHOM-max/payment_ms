package application

import (
	"context"
	"errors"

	"payment_ms/domain"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// validTransitions defines allowed next states from a given status.
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

	allowed, ok := validTransitions[payment.Status]
	if !ok || !allowed[input.Status] {
		return nil, ErrInvalidStatusTransition
	}

	if err := uc.repo.UpdateStatus(ctx, input.PaymentID, input.Status); err != nil {
		return nil, err
	}

	if err := uc.signaler.Signal(ctx, payment.WorkflowID, input.Status, input.PaymentID); err != nil {
		return nil, err
	}

	return &HandleWebhookOutput{
		WorkflowID: payment.WorkflowID,
		Status:     input.Status,
	}, nil
}
