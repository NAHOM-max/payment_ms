package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"payment_ms/application"
	"payment_ms/domain"
)

type DeliveryConfirmedConsumer struct {
	reader *kafkago.Reader
	uc     *application.HandleDeliveryConfirmedUseCase
}

func NewDeliveryConfirmedConsumer(brokers []string, uc *application.HandleDeliveryConfirmedUseCase) *DeliveryConfirmedConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   "delivery.confirmed",
		GroupID: "payment-service-group",
	})
	return &DeliveryConfirmedConsumer{reader: reader, uc: uc}
}

func (c *DeliveryConfirmedConsumer) Run(ctx context.Context) error {
	defer c.reader.Close()
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		eventID := string(msg.Key)
		log.Printf("event received event_id=%s topic=%s", eventID, msg.Topic)

		var event domain.DeliveryConfirmedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("deserialize failed event_id=%s: %v", eventID, err)
			// commit malformed messages to avoid infinite retry
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := c.uc.Execute(ctx, eventID, event, msg.Value); err != nil {
			log.Printf("handle failed event_id=%s: %v — will retry", eventID, err)
			// do not commit; offset stays, message will be redelivered
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit failed event_id=%s: %v", eventID, err)
		}
	}
}
