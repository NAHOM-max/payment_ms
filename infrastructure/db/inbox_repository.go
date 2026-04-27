package db

import (
	"context"
	"time"

	prisma "payment_ms/infrastructure/db/prisma"

	"payment_ms/domain"
)

type InboxRepository struct {
	client *prisma.PrismaClient
}

func NewInboxRepository(client *prisma.PrismaClient) *InboxRepository {
	return &InboxRepository{client: client}
}

func (r *InboxRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	row, err := r.client.InboxEvent.FindUnique(
		prisma.InboxEvent.EventID.Equals(eventID),
	).Exec(ctx)
	if err != nil {
		return false, nil // not found → not seen yet
	}
	return row.Processed, nil
}

func (r *InboxRepository) Save(ctx context.Context, event domain.InboxEvent) error {
	_, err := r.client.InboxEvent.CreateOne(
		prisma.InboxEvent.EventID.Set(event.EventID),
		prisma.InboxEvent.EventType.Set(event.EventType),
		prisma.InboxEvent.Payload.Set(event.Payload),
	).Exec(ctx)
	return err
}

func (r *InboxRepository) MarkProcessed(ctx context.Context, eventID string) error {
	_, err := r.client.InboxEvent.FindUnique(
		prisma.InboxEvent.EventID.Equals(eventID),
	).Update(
		prisma.InboxEvent.Processed.Set(true),
		prisma.InboxEvent.UpdatedAt.Set(time.Now().UTC()),
	).Exec(ctx)
	return err
}
