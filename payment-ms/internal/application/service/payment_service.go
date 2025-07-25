package service

import (
	"fmt"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/external"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

// TODO: change currency and makeing chapa suitable

type PaymentServiceImpl struct {
	repo           domain.PaymentRepository
	nats           *messaging.PaymentEventPublisher
	paymentGateway external.PaymentGateway
}

func NewPaymentService(repo domain.PaymentRepository, nats *messaging.PaymentEventPublisher, pg external.PaymentGateway) domain.PaymentService {
	service := &PaymentServiceImpl{
		repo:           repo,
		nats:           nats,
		paymentGateway: pg,
	}
	return service
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
		s.nats.PublishPaymentProcessed(payment, false, err.Error())
		return err
	}

	// Update payment with transaction ID and success status
	payment.TransactionID = transactionID
	payment.Status = domain.PaymentStatusCompleted
	s.repo.Update(payment.ID.Hex(), payment)

	// Publish payment success event
	s.nats.PublishPaymentProcessed(payment, true, "")

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
	s.nats.PublishPaymentRefunded(payment, payment.Amount)
	return nil
}

func (s *PaymentServiceImpl) ListPayments(limit, offset int) ([]*domain.Payment, error) {
	return s.repo.List(limit, offset)
}
