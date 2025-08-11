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
	UserCreatedEvent         = "user.created"
	UserDeletedEvent         = "user.deleted"
	UserUpdatedEvent         = "user.updated"
	ProductCreatedEvent      = "product.created"
	ProductUpdatedEvent      = "product.updated"
	ProductStockUpdatedEvent = "product.stock.updated"
	OrderCreatedEvent        = "order.created"
	OrderCancelledEvent      = "order.cancelled"
	OrderUpdatedEvent        = "order.updated"
	PaymentProcessedEvent    = "payment.processed"
	StockCheckEvent          = "stock.check"
	StockReserveEvent        = "stock.reserve"
	StockCheckResponseEvent  = "stock.check.response"
)
