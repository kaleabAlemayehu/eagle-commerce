package service

import (
	"encoding/json"
	"fmt"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/external"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

// TODO: change currency and makeing chapa suitable

type PaymentServiceImpl struct {
	repo           domain.PaymentRepository
	natsClient     *messaging.NATSClient
	paymentGateway external.PaymentGateway
}

func NewPaymentService(repo domain.PaymentRepository, natsClient *messaging.NATSClient, pg external.PaymentGateway) domain.PaymentService {
	service := &PaymentServiceImpl{
		repo:           repo,
		natsClient:     natsClient,
		paymentGateway: pg,
	}

	// Subscribe to order events
	service.subscribeToEvents()

	return service
}

func (s *PaymentServiceImpl) subscribeToEvents() {
	s.natsClient.Subscribe("order.created", s.handleOrderCreated)
}

func (s *PaymentServiceImpl) handleOrderCreated(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		return
	}

	total, ok := event.Data["total"].(float64)
	if !ok {
		return
	}

	// Auto-create payment record
	payment := &domain.Payment{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   total,
		Currency: "USD",
		Method:   domain.PaymentMethodCard, // Default method
	}

	s.repo.Create(payment)
}

func (s *PaymentServiceImpl) ProcessPayment(payment *domain.Payment) error {
	if err := utils.ValidateStruct(payment); err != nil {
		return err
	}

	// Save payment record
	if err := s.repo.Create(payment); err != nil {
		return err
	}

	// Process payment through gateway
	transactionID, err := s.paymentGateway.ProcessPayment(payment)
	if err != nil {
		// Update payment status to failed
		s.repo.UpdateStatus(payment.ID.Hex(), domain.PaymentStatusFailed)

		// Publish payment failed event
		event := models.Event{
			Type:   models.PaymentProcessedEvent,
			Source: "payment-service",
			Data: map[string]interface{}{
				"payment_id": payment.ID.Hex(),
				"order_id":   payment.OrderID,
				"status":     "failed",
				"error":      err.Error(),
			},
		}
		s.natsClient.Publish("payment.processed", event)

		return err
	}

	// Update payment with transaction ID and success status
	payment.TransactionID = transactionID
	payment.Status = domain.PaymentStatusCompleted
	s.repo.Update(payment.ID.Hex(), payment)

	// Publish payment success event
	event := models.Event{
		Type:   models.PaymentProcessedEvent,
		Source: "payment-service",
		Data: map[string]interface{}{
			"payment_id":     payment.ID.Hex(),
			"order_id":       payment.OrderID,
			"transaction_id": transactionID,
			"status":         "completed",
			"amount":         payment.Amount,
		},
	}
	s.natsClient.Publish("payment.processed", event)

	return nil
}

func (s *PaymentServiceImpl) GetPayment(id string) (*domain.Payment, error) {
	return s.repo.GetByID(id)
}

func (s *PaymentServiceImpl) GetPaymentByOrder(orderID string) (*domain.Payment, error) {
	return s.repo.GetByOrderID(orderID)
}

func (s *PaymentServiceImpl) RefundPayment(id string) error {
	payment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if payment.Status != domain.PaymentStatusCompleted {
		return fmt.Errorf("cannot refund payment with status: %s", payment.Status)
	}

	// Process refund through gateway
	if err := s.paymentGateway.RefundPayment(payment.TransactionID, payment.Amount); err != nil {
		return err
	}

	// Update payment status
	if err := s.repo.UpdateStatus(id, domain.PaymentStatusRefunded); err != nil {
		return err
	}

	// Publish refund event
	event := models.Event{
		Type:   "payment.refunded",
		Source: "payment-service",
		Data: map[string]interface{}{
			"payment_id": id,
			"order_id":   payment.OrderID,
			"amount":     payment.Amount,
		},
	}
	s.natsClient.Publish("payment.refunded", event)

	return nil
}

func (s *PaymentServiceImpl) ListPayments(limit, offset int) ([]*domain.Payment, error) {
	return s.repo.List(limit, offset)
}
