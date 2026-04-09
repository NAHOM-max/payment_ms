package db

import (
	"context"
	"time"

	prisma "payment_ms/infrastructure/db/prisma"

	"payment_ms/domain"
)

type PaymentRepository struct {
	client *prisma.PrismaClient
}

func NewPaymentRepository(client *prisma.PrismaClient) *PaymentRepository {
	return &PaymentRepository{client: client}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.client.Payment.CreateOne(
		prisma.Payment.CustomerID.Set(p.CustomerID),
		prisma.Payment.OrderID.Set(p.OrderID),
		prisma.Payment.WorkflowID.Set(p.WorkflowID),
		prisma.Payment.Amount.Set(p.Amount),
		prisma.Payment.Status.Set(toDBStatus(p.Status)),
		prisma.Payment.CreatedAt.Set(p.CreatedAt),
		prisma.Payment.UpdatedAt.Set(p.UpdatedAt),
		prisma.Payment.ID.Set(p.ID),
	).Exec(ctx)
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, paymentID string) (*domain.Payment, error) {
	row, err := r.client.Payment.FindUnique(
		prisma.Payment.ID.Equals(paymentID),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return toDomain(row), nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, paymentID string, status domain.PaymentStatus) error {
	_, err := r.client.Payment.FindUnique(
		prisma.Payment.ID.Equals(paymentID),
	).Update(
		prisma.Payment.Status.Set(toDBStatus(status)),
		prisma.Payment.UpdatedAt.Set(time.Now().UTC()),
	).Exec(ctx)
	return err
}

func (r *PaymentRepository) UpdateStatusConditional(ctx context.Context, paymentID string, fromStatus, toStatus domain.PaymentStatus) (bool, error) {
	result, err := r.client.Payment.FindMany(
		prisma.Payment.ID.Equals(paymentID),
		prisma.Payment.Status.Equals(toDBStatus(fromStatus)),
	).Update(
		prisma.Payment.Status.Set(toDBStatus(toStatus)),
		prisma.Payment.UpdatedAt.Set(time.Now().UTC()),
	).Exec(ctx)
	if err != nil {
		return false, err
	}
	return result.Count > 0, nil
}

// ── Mapping ───────────────────────────────────────────────────────────────────

func toDomain(row *prisma.PaymentModel) *domain.Payment {
	return &domain.Payment{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		OrderID:    row.OrderID,
		WorkflowID: row.WorkflowID,
		Amount:     row.Amount,
		Status:     domain.PaymentStatus(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func toDBStatus(s domain.PaymentStatus) prisma.PaymentStatus {
	return prisma.PaymentStatus(s)
}
