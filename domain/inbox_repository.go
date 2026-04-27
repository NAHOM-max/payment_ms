package domain

import "context"

type InboxRepository interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	Save(ctx context.Context, event InboxEvent) error
	MarkProcessed(ctx context.Context, eventID string) error
}
