package domain

import "context"

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, paymentID string) (*Payment, error)
	UpdateStatus(ctx context.Context, paymentID string, status PaymentStatus) error
}
