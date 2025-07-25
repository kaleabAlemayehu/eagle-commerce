// services/payment-ms/internal/infrastructure/messaging/event_handlers.go
package messaging

import (
	"encoding/json"
	"log"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type PaymentEventHandler struct {
	paymentService domain.PaymentService
	natsClient     *messaging.NATSClient
	publisher      *PaymentEventPublisher
}

func NewPaymentEventHandler(paymentService domain.PaymentService, natsClient *messaging.NATSClient) *PaymentEventHandler {
	return &PaymentEventHandler{
		paymentService: paymentService,
		natsClient:     natsClient,
		publisher:      NewPaymentEventPublisher(natsClient),
	}
}

func (h *PaymentEventHandler) StartListening() error {
	// Subscribe to order events
	_, err := h.natsClient.Subscribe("order.created", h.handleOrderCreated)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe("order.cancelled", h.handleOrderCancelled)
	if err != nil {
		return err
	}

	// Subscribe to refund requests
	_, err = h.natsClient.Subscribe("refund.requested", h.handleRefundRequested)
	return err
}

func (h *PaymentEventHandler) handleOrderCreated(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling order.created event: %v", err)
		return
	}

	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		log.Printf("Invalid order_id in order.created event")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Invalid user_id in order.created event")
		return
	}

	total, ok := event.Data["total"].(float64)
	if !ok {
		log.Printf("Invalid total in order.created event")
		return
	}

	// Create a pending payment record
	payment := &domain.Payment{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   total,
		Currency: "USD",                    // Default currency
		Method:   domain.PaymentMethodCard, // Default method
	}

	if err := h.paymentService.ProcessPayment(payment); err != nil {
		log.Printf("Error creating payment record for order %s: %v", orderID, err)

		// Publish payment failed event
		h.publisher.PublishPaymentFailed(orderID, userID, total, err.Error())
		return
	}

	log.Printf("Payment record created for order %s", orderID)
}

func (h *PaymentEventHandler) handleOrderCancelled(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling order.cancelled event: %v", err)
		return
	}

	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		log.Printf("Invalid order_id in order.cancelled event")
		return
	}

	// Check if there's a completed payment for this order
	payment, err := h.paymentService.GetPaymentByOrder(orderID)
	if err != nil {
		log.Printf("No payment found for cancelled order %s", orderID)
		return
	}

	// If payment was completed, initiate refund
	if payment.Status == domain.PaymentStatusCompleted {
		if err := h.paymentService.RefundPayment(payment.ID.Hex()); err != nil {
			log.Printf("Error refunding payment %s for cancelled order %s: %v",
				payment.ID.Hex(), orderID, err)
			return
		}

		log.Printf("Refund initiated for payment %s (order %s)", payment.ID.Hex(), orderID)
	}
}

func (h *PaymentEventHandler) handleRefundRequested(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling refund.requested event: %v", err)
		return
	}

	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		log.Printf("Invalid payment_id in refund.requested event")
		return
	}

	if err := h.paymentService.RefundPayment(paymentID); err != nil {
		log.Printf("Error processing refund for payment %s: %v", paymentID, err)
		return
	}

	log.Printf("Refund processed for payment %s", paymentID)
}
