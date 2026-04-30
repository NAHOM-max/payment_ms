package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"payment_ms/application"
	"payment_ms/domain"
)

const serviceName = "payment-service"

type DeliveryConfirmedConsumer struct {
	reader     *kafkago.Reader
	uc         *application.HandleDeliveryConfirmedUseCase
	dlq        domain.DLQProducer
	maxRetries int
}

func NewDeliveryConfirmedConsumer(
	brokers []string,
	uc *application.HandleDeliveryConfirmedUseCase,
	dlq domain.DLQProducer,
	maxRetries int,
) *DeliveryConfirmedConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   "delivery.confirmed",
		GroupID: "payment-service-group",
	})
	return &DeliveryConfirmedConsumer{reader: reader, uc: uc, dlq: dlq, maxRetries: maxRetries}
}

func (c *DeliveryConfirmedConsumer) Run(ctx context.Context) error {
	defer c.reader.Close()
	defer func() {
		if err := c.dlq.Close(); err != nil {
			log.Printf("dlq close error: %v", err)
		}
	}()
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
			log.Printf("deserialize failed event_id=%s: %v — sending to DLQ", eventID, err)
			c.sendToDLQ(ctx, eventID, event, msg.Value, err, 0)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		execErr := c.executeWithRetry(ctx, eventID, event, msg.Value)
		if execErr != nil {
			log.Printf("sending event to DLQ event_id=%s order_id=%s shipment_id=%s",
				eventID, event.OrderID, event.ShipmentID)
			c.sendToDLQ(ctx, eventID, event, msg.Value, execErr, c.maxRetries)
		}

		// always commit — either processed successfully or handed off to DLQ
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit failed event_id=%s: %v", eventID, err)
		}
	}
}

func (c *DeliveryConfirmedConsumer) executeWithRetry(ctx context.Context, eventID string, event domain.DeliveryConfirmedEvent, raw []byte) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("retry attempt %d for event_id=%s order_id=%s shipment_id=%s",
				attempt, eventID, event.OrderID, event.ShipmentID)
		}

		lastErr = c.uc.Execute(ctx, eventID, event, raw)
		if lastErr == nil {
			return nil
		}

		if errors.Is(lastErr, domain.ErrNonRetryable) {
			log.Printf("non-retryable error event_id=%s: %v — skipping retries", eventID, lastErr)
			return lastErr
		}

		if attempt < c.maxRetries {
			time.Sleep(backoff(attempt))
		}
	}
	return lastErr
}

func (c *DeliveryConfirmedConsumer) sendToDLQ(ctx context.Context, eventID string, event domain.DeliveryConfirmedEvent, _ []byte, cause error, retryCount int) {
	dlqEvent := domain.DLQEvent{
		OriginalEvent: event,
		Error:         cause.Error(),
		Service:       serviceName,
		RetryCount:    retryCount,
		FailedAt:      time.Now().UTC(),
	}
	if err := c.dlq.Send(ctx, eventID, dlqEvent); err != nil {
		log.Printf("DLQ publish failed event_id=%s: %v", eventID, err)
	}
}

// backoff returns a simple exponential delay: 1s, 2s, 4s, …
func backoff(attempt int) time.Duration {
	d := time.Duration(1<<attempt) * time.Second
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}
