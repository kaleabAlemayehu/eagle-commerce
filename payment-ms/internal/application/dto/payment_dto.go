package dto

import "time"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ProcessPaymentRequest struct {
	OrderID     string       `json:"order_id" validate:"required"`
	UserID      string       `json:"user_id" validate:"required"`
	Amount      float64      `json:"amount" validate:"gt=0"`
	Currency    string       `json:"currency" validate:"required"`
	Method      string       `json:"method" validate:"required,oneof=card paypal bank"`
	CardDetails *CardDetails `json:"card_details,omitempty"`
}

type CardDetails struct {
	Number      string `json:"number" validate:"required"`
	ExpiryMonth int    `json:"expiry_month" validate:"required,min=1,max=12"`
	ExpiryYear  int    `json:"expiry_year" validate:"required"`
	CVV         string `json:"cvv" validate:"required,len=3"`
	HolderName  string `json:"holder_name" validate:"required"`
}

type PaymentResponse struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Method        string    `json:"method"`
	TransactionID string    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RefundPaymentRequest struct {
	Reason string   `json:"reason,omitempty"`
	Amount *float64 `json:"amount,omitempty"` // Partial refund amount
}

type RefundResponse struct {
	PaymentID    string    `json:"payment_id"`
	RefundAmount float64   `json:"refund_amount"`
	RefundID     string    `json:"refund_id"`
	Status       string    `json:"status"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type PaymentListResponse struct {
	Payments   []PaymentResponse `json:"payments"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

type PaymentSummaryResponse struct {
	TotalPayments      int     `json:"total_payments"`
	TotalAmount        float64 `json:"total_amount"`
	SuccessfulPayments int     `json:"successful_payments"`
	FailedPayments     int     `json:"failed_payments"`
	RefundedPayments   int     `json:"refunded_payments"`
	TotalRefunds       float64 `json:"total_refunds"`
}

type WebhookPayload struct {
	Event         string                 `json:"event"`
	TransactionID string                 `json:"transaction_id"`
	PaymentID     string                 `json:"payment_id"`
	Status        string                 `json:"status"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
}
