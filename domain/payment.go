package domain

import (
	"time"
)

type PaymentStatus string

const (
	StatusInitiated  PaymentStatus = "INITIATED"
	StatusSuccessful PaymentStatus = "SUCCESSFUL"
	StatusFailed     PaymentStatus = "FAILED"
	StatusCanceled   PaymentStatus = "CANCELED"
)

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
