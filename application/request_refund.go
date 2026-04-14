package application

import (
	"context"
	"errors"

	"payment_ms/domain"
)

type RequestRefundOutput struct {
	PaymentID  string
	WorkflowID string
	Status     domain.PaymentStatus
}

type RequestRefundUseCase struct {
	repo domain.PaymentRepository
}

func NewRequestRefundUseCase(repo domain.PaymentRepository) *RequestRefundUseCase {
	return &RequestRefundUseCase{repo: repo}
}

func (uc *RequestRefundUseCase) Execute(ctx context.Context, paymentID string) (*RequestRefundOutput, error) {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	if err := payment.RequestRefund(); err != nil {
		if errors.Is(err, domain.ErrRefundNotAllowed) {
			return nil, ErrInvalidStatusTransition
		}
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, payment.ID, payment.Status); err != nil {
		return nil, err
	}

	return &RequestRefundOutput{
		PaymentID:  payment.ID,
		WorkflowID: payment.WorkflowID,
		Status:     payment.Status,
	}, nil
}
