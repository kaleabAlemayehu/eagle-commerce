package external

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
)

type PaymentGateway interface {
	ProcessPayment(payment *domain.Payment) (string, error)
	RefundPayment(transactionID string, amount float64) error
}

// TODO: replace this with chapa payment integration but for now
// MockPaymentGateway simulates a payment gateway
type MockPaymentGateway struct{}

func NewMockPaymentGateway() PaymentGateway {
	return &MockPaymentGateway{}
}

func (pg *MockPaymentGateway) ProcessPayment(payment *domain.Payment) (string, error) {
	// Simulate processing time
	time.Sleep(1 * time.Second)

	// Generate random transaction ID
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	transactionID := fmt.Sprintf("TXN_%d_%d", time.Now().Unix(), n.Int64())

	// Simulate success/failure (90% success rate)
	success, _ := rand.Int(rand.Reader, big.NewInt(10))
	if success.Int64() < 9 {
		return transactionID, nil
	}

	return "", fmt.Errorf("payment processing failed")
}

func (pg *MockPaymentGateway) RefundPayment(transactionID string, amount float64) error {
	// Simulate refund processing
	time.Sleep(500 * time.Millisecond)
	return nil
}
