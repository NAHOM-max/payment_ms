package application

import (
	"context"
	"encoding/json"
	"log"

	"payment_ms/domain"
)

type HandleDeliveryConfirmedUseCase struct {
	inbox domain.InboxRepository
}

func NewHandleDeliveryConfirmedUseCase(inbox domain.InboxRepository) *HandleDeliveryConfirmedUseCase {
	return &HandleDeliveryConfirmedUseCase{inbox: inbox}
}

func (uc *HandleDeliveryConfirmedUseCase) Execute(ctx context.Context, eventID string, event domain.DeliveryConfirmedEvent, rawPayload []byte) error {
	processed, err := uc.inbox.Exists(ctx, eventID)
	if err != nil {
		return err
	}
	if processed {
		log.Printf("duplicate skipped event_id=%s order_id=%s shipment_id=%s", eventID, event.OrderID, event.ShipmentID)
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := uc.inbox.Save(ctx, domain.InboxEvent{
		EventID:   eventID,
		EventType: "delivery.confirmed",
		Payload:   payload,
	}); err != nil {
		return err
	}

	log.Printf("Payment service received delivery.confirmed for order %s event_id=%s shipment_id=%s", event.OrderID, eventID, event.ShipmentID)

	if err := uc.inbox.MarkProcessed(ctx, eventID); err != nil {
		return err
	}

	log.Printf("processed event_id=%s order_id=%s shipment_id=%s", eventID, event.OrderID, event.ShipmentID)
	return nil
}
