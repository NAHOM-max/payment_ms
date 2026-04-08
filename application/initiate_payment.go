package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"payment_ms/domain"
)

type InitiatePaymentInput struct {
	CustomerID string
	OrderID    string
	WorkflowID string
	Amount     float64
}

type InitiatePaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewInitiatePaymentUseCase(repo domain.PaymentRepository) *InitiatePaymentUseCase {
	return &InitiatePaymentUseCase{repo: repo}
}

func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, input InitiatePaymentInput) (string, error) {
	now := time.Now().UTC()
	payment := &domain.Payment{
		ID:         uuid.NewString(),
		CustomerID: input.CustomerID,
		OrderID:    input.OrderID,
		WorkflowID: input.WorkflowID,
		Amount:     input.Amount,
		Status:     domain.StatusInitiated,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return "", err
	}

	return payment.ID, nil
}
