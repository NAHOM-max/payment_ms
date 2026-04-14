package domain

import (
	"errors"
	"time"
)

type PaymentStatus string

const (
	StatusInitiated       PaymentStatus = "INITIATED"
	StatusSuccessful      PaymentStatus = "SUCCESSFUL"
	StatusFailed          PaymentStatus = "FAILED"
	StatusCanceled        PaymentStatus = "CANCELED"
	StatusRefundRequested PaymentStatus = "REFUND_REQUESTED"
)

var ErrRefundNotAllowed = errors.New("refund only allowed for successful payments")

type Payment struct {
	ID         string
	CustomerID string
	OrderID    string
	WorkflowID string
	Amount     float64
	Status     PaymentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p *Payment) RequestRefund() error {
	if p.Status != StatusSuccessful {
		return ErrRefundNotAllowed
	}
	p.Status = StatusRefundRequested
	p.UpdatedAt = time.Now().UTC()
	return nil
}
