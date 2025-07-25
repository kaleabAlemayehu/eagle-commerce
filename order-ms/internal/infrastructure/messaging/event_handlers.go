package messaging

import (
	"encoding/json"
	"log"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type OrderEventHandler struct {
	orderService domain.OrderService
	natsClient   *messaging.NATSClient
	publisher    *OrderEventPublisher
}

func NewOrderEventHandler(orderService domain.OrderService, natsClient *messaging.NATSClient) *OrderEventHandler {
	return &OrderEventHandler{
		orderService: orderService,
		natsClient:   natsClient,
		publisher:    NewOrderEventPublisher(natsClient),
	}
}

func (h *OrderEventHandler) StartListening() error {
	// Subscribe to payment events
	_, err := h.natsClient.Subscribe("payment.processed", h.handlePaymentProcessed)
	if err != nil {
		return err
	}

	// Subscribe to inventory events
	_, err = h.natsClient.Subscribe("stock.check.response", h.handleStockCheckResponse)
	if err != nil {
		return err
	}

	// Subscribe to user events
	_, err = h.natsClient.Subscribe("user.updated", h.handleUserUpdated)
	return err
}

func (h *OrderEventHandler) handlePaymentProcessed(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling payment.processed event: %v", err)
		return
	}

	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		log.Printf("Invalid order_id in payment.processed event")
		return
	}

	status, ok := event.Data["status"].(string)
	if !ok {
		log.Printf("Invalid status in payment.processed event")
		return
	}

	// Update order status based on payment result
	var newStatus domain.OrderStatus
	switch status {
	case "completed":
		newStatus = domain.OrderStatusConfirmed
	case "failed":
		newStatus = domain.OrderStatusCancelled
	default:
		log.Printf("Unknown payment status: %s", status)
		return
	}

	if err := h.orderService.UpdateOrderStatus(orderID, newStatus); err != nil {
		log.Printf("Error updating order status: %v", err)
		return
	}

	log.Printf("Order %s status updated to %s based on payment result", orderID, newStatus)

	// If order was cancelled due to payment failure, publish cancellation event
	if newStatus == domain.OrderStatusCancelled {
		order, err := h.orderService.GetOrder(orderID)
		if err == nil {
			h.publisher.PublishOrderCancelled(order)
		}
	}
}

func (h *OrderEventHandler) handleStockCheckResponse(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling stock.check.response event: %v", err)
		return
	}

	productID, ok := event.Data["product_id"].(string)
	if !ok {
		log.Printf("Invalid product_id in stock.check.response event")
		return
	}

	available, ok := event.Data["available"].(bool)
	if !ok {
		log.Printf("Invalid available in stock.check.response event")
		return
	}

	requestID, ok := event.Data["request_id"].(string)
	if !ok {
		log.Printf("Invalid request_id in stock.check.response event")
		return
	}

	log.Printf("Stock check response for product %s (request %s): available=%t",
		productID, requestID, available)

	// INFO:
	// Here you could store the response for correlation with pending orders
	// or handle insufficient stock scenarios
}

func (h *OrderEventHandler) handleUserUpdated(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling user.updated event: %v", err)
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Invalid user_id in user.updated event")
		return
	}

	log.Printf("User %s was updated - orders may need address updates", userID)
	// Here you could implement logic to update user information in orders
	// or send notifications about profile changes
}
