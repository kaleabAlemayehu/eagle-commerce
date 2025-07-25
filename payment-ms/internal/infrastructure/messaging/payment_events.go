// services/payment-ms/internal/infrastructure/messaging/payment_events.go
package messaging

import (
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type PaymentEventPublisher struct {
	natsClient *messaging.NATSClient
}

func NewPaymentEventPublisher(natsClient *messaging.NATSClient) *PaymentEventPublisher {
	return &PaymentEventPublisher{
		natsClient: natsClient,
	}
}

func (p *PaymentEventPublisher) PublishPaymentProcessed(payment *domain.Payment, success bool, errorMsg string) error {
	status := "completed"
	if !success {
		status = "failed"
	}

	eventData := map[string]interface{}{
		"payment_id": payment.ID.Hex(),
		"order_id":   payment.OrderID,
		"user_id":    payment.UserID,
		"amount":     payment.Amount,
		"currency":   payment.Currency,
		"method":     string(payment.Method),
		"status":     status,
	}

	if success && payment.TransactionID != "" {
		eventData["transaction_id"] = payment.TransactionID
	}

	if !success && errorMsg != "" {
		eventData["error"] = errorMsg
	}

	event := models.Event{
		ID:        generateEventID(),
		Type:      models.PaymentProcessedEvent,
		Source:    "payment-service",
		Data:      eventData,
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("payment.processed", event)
}

func (p *PaymentEventPublisher) PublishPaymentRefunded(payment *domain.Payment, refundAmount float64) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   "payment.refunded",
		Source: "payment-service",
		Data: map[string]interface{}{
			"payment_id":      payment.ID.Hex(),
			"order_id":        payment.OrderID,
			"user_id":         payment.UserID,
			"original_amount": payment.Amount,
			"refund_amount":   refundAmount,
			"refunded_at":     time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("payment.refunded", event)
}

func (p *PaymentEventPublisher) PublishPaymentFailed(orderID, userID string, amount float64, reason string) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   "payment.failed",
		Source: "payment-service",
		Data: map[string]interface{}{
			"order_id":  orderID,
			"user_id":   userID,
			"amount":    amount,
			"reason":    reason,
			"failed_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("payment.failed", event)
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + "payment"
}
