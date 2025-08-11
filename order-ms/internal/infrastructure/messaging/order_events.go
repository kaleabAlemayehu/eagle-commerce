package messaging

import (
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"

	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type OrderEventPublisher struct {
	natsClient *messaging.NATSClient
}

func NewOrderEventPublisher(natsClient *messaging.NATSClient) *OrderEventPublisher {
	return &OrderEventPublisher{
		natsClient: natsClient,
	}
}

func (p *OrderEventPublisher) PublishOrderCreated(order *domain.Order) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.OrderCreatedEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id":   order.ID.Hex(),
			"user_id":    order.UserID,
			"total":      order.Total,
			"status":     string(order.Status),
			"items":      order.Items,
			"address":    order.Address,
			"created_at": order.CreatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.OrderCreatedEvent, event)
}

func (p *OrderEventPublisher) PublishOrderUpdated(order *domain.Order, oldStatus domain.OrderStatus) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.OrderUpdatedEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id":   order.ID.Hex(),
			"user_id":    order.UserID,
			"old_status": string(oldStatus),
			"new_status": string(order.Status),
			"total":      order.Total,
			"updated_at": order.UpdatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.OrderUpdatedEvent, event)
}

func (p *OrderEventPublisher) PublishOrderCancelled(order *domain.Order) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.OrderCancelledEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id":     order.ID.Hex(),
			"user_id":      order.UserID,
			"total":        order.Total,
			"items":        order.Items,
			"cancelled_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.OrderCancelledEvent, event)
}

func (p *OrderEventPublisher) PublishOrderShipped(order *domain.Order, trackingID string) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.OrderShippedEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id":    order.ID.Hex(),
			"user_id":     order.UserID,
			"tracking_id": trackingID,
			"address":     order.Address,
			"shipped_at":  time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.OrderShippedEvent, event)
}

func (p *OrderEventPublisher) PublishStockCheck(item *domain.OrderItem) error {
	// Publish stock check request
	event := models.Event{
		Type:   models.StockCheckEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
		},
	}
	return p.natsClient.Publish(models.StockCheckEvent, event)
}

func (p *OrderEventPublisher) PublishReserveStock(item *domain.OrderItem) error {
	// Publish stock check request
	event := models.Event{
		Type:   models.StockReserveEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
		},
	}
	return p.natsClient.Publish(models.StockReserveEvent, event)
}
