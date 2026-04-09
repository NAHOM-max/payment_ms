package domain

import "context"

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, paymentID string) (*Payment, error)
	UpdateStatus(ctx context.Context, paymentID string, status PaymentStatus) error
	// UpdateStatusConditional updates status only if the current DB status matches
	// fromStatus. Returns true if the row was updated, false if the condition did
	// not match (concurrent update already applied).
	UpdateStatusConditional(ctx context.Context, paymentID string, fromStatus, toStatus PaymentStatus) (bool, error)
}
