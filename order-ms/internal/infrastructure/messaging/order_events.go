// services/order-ms/internal/infrastructure/messaging/order_events.go
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
		ID:     generateEventID(),
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

	return p.natsClient.Publish("order.created", event)
}

func (p *OrderEventPublisher) PublishOrderUpdated(order *domain.Order, oldStatus domain.OrderStatus) error {
	event := models.Event{
		ID:     generateEventID(),
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

	return p.natsClient.Publish("order.updated", event)
}

func (p *OrderEventPublisher) PublishOrderCancelled(order *domain.Order) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   "order.cancelled",
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

	return p.natsClient.Publish("order.cancelled", event)
}

func (p *OrderEventPublisher) PublishOrderShipped(order *domain.Order, trackingID string) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   "order.shipped",
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

	return p.natsClient.Publish("order.shipped", event)
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + "order"
}
