package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"payment_ms/domain"
)

const dlqTopic = "delivery.confirmed.dlq"

type KafkaDLQProducer struct {
	writer *kafkago.Writer
}

func NewKafkaDLQProducer(brokers []string) *KafkaDLQProducer {
	return &KafkaDLQProducer{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(brokers...),
			Topic:    dlqTopic,
			Balancer: &kafkago.LeastBytes{},
		},
	}
}

func (p *KafkaDLQProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaDLQProducer) Send(ctx context.Context, key string, event domain.DLQEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: payload,
	})
	if err != nil {
		log.Printf("DLQ publish failed event_id=%s order_id=%s shipment_id=%s: %v",
			key, event.OriginalEvent.OrderID, event.OriginalEvent.ShipmentID, err)
		return err
	}

	log.Printf("DLQ publish success event_id=%s order_id=%s shipment_id=%s retries=%d",
		key, event.OriginalEvent.OrderID, event.OriginalEvent.ShipmentID, event.RetryCount)
	return nil
}
