// services/order-ms/internal/application/dto/order_dto.go
package dto

import "time"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type CreateOrderRequest struct {
	UserID  string                   `json:"user_id" validate:"required"`
	Items   []CreateOrderItemRequest `json:"items" validate:"required,dive"`
	Address AddressRequest           `json:"address" validate:"required"`
}

type CreateOrderItemRequest struct {
	ProductID string  `json:"product_id" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Price     float64 `json:"price" validate:"gt=0"`
	Quantity  int     `json:"quantity" validate:"gt=0"`
}

type AddressRequest struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	State   string `json:"state" validate:"required"`
	ZipCode string `json:"zip_code" validate:"required"`
	Country string `json:"country" validate:"required"`
}

type OrderResponse struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	Items     []OrderItemResponse `json:"items"`
	Total     float64             `json:"total"`
	Status    string              `json:"status"`
	Address   AddressResponse     `json:"address"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

type AddressResponse struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed shipped delivered cancelled"`
}

type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

type OrderSummaryResponse struct {
	TotalOrders     int     `json:"total_orders"`
	TotalValue      float64 `json:"total_value"`
	PendingOrders   int     `json:"pending_orders"`
	CompletedOrders int     `json:"completed_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
}

type TrackingResponse struct {
	OrderID       string         `json:"order_id"`
	Status        string         `json:"status"`
	TrackingID    string         `json:"tracking_id,omitempty"`
	StatusHistory []StatusChange `json:"status_history"`
}

type StatusChange struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Note      string    `json:"note,omitempty"`
}
