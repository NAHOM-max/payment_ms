package domain

import "time"

type InboxEvent struct {
	ID        string
	EventID   string
	EventType string
	Payload   []byte
	Processed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeliveryConfirmedEvent struct {
	ShipmentID     string    `json:"shipment_id"`
	OrderID        string    `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	DeliveredAt    time.Time `json:"delivered_at"`
}
