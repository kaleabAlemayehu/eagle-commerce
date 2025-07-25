package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Payment struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderID       string             `json:"order_id" bson:"order_id" validate:"required"`
	UserID        string             `json:"user_id" bson:"user_id" validate:"required"`
	Amount        float64            `json:"amount" bson:"amount" validate:"gt=0"`
	Currency      string             `json:"currency" bson:"currency" validate:"required"`
	Status        PaymentStatus      `json:"status" bson:"status"`
	Method        PaymentMethod      `json:"method" bson:"method" validate:"required"`
	TransactionID string             `json:"transaction_id" bson:"transaction_id"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodPaypal PaymentMethod = "paypal"
	PaymentMethodBank   PaymentMethod = "bank"
)

type PaymentRepository interface {
	Create(payment *Payment) error
	GetByID(id string) (*Payment, error)
	GetByOrderID(orderID string) (*Payment, error)
	Update(id string, payment *Payment) error
	UpdateStatus(id string, status PaymentStatus) error
	List(limit, offset int) ([]*Payment, error)
}

type PaymentService interface {
	ProcessPayment(payment *Payment) error
	GetPayment(id string) (*Payment, error)
	GetPaymentByOrder(orderID string) (*Payment, error)
	RefundPayment(id string) error
	ListPayments(limit, offset int) ([]*Payment, error)
}
