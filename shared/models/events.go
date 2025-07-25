package models

import "time"

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Event types
const (
	UserCreatedEvent      = "user.created"
	UserUpdatedEvent      = "user.updated"
	ProductCreatedEvent   = "product.created"
	ProductUpdatedEvent   = "product.updated"
	OrderCreatedEvent     = "order.created"
	OrderUpdatedEvent     = "order.updated"
	PaymentProcessedEvent = "payment.processed"
)
