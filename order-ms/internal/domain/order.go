package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id" validate:"required"`
	Items     []OrderItem        `json:"items" bson:"items" validate:"required,dive"`
	Total     float64            `json:"total" bson:"total"`
	Status    OrderStatus        `json:"status" bson:"status"`
	Address   Address            `json:"address" bson:"address" validate:"required"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id" bson:"product_id" validate:"required"`
	Name      string  `json:"name" bson:"name"`
	Price     float64 `json:"price" bson:"price" validate:"gt=0"`
	Quantity  int     `json:"quantity" bson:"quantity" validate:"gt=0"`
}

type Address struct {
	Street  string `json:"street" bson:"street" validate:"required"`
	City    string `json:"city" bson:"city" validate:"required"`
	State   string `json:"state" bson:"state" validate:"required"`
	ZipCode string `json:"zip_code" bson:"zip_code" validate:"required"`
	Country string `json:"country" bson:"country" validate:"required"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderRepository interface {
	Create(ctx context.Context, order *Order) (*Order, error)
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error)
	Update(ctx context.Context, id string, order *Order) (*Order, error)
	UpdateStatus(ctx context.Context, id string, currentStatus OrderStatus, newStatus OrderStatus) (*Order, error)
	List(ctx context.Context, limit, offset int) ([]*Order, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, id string) (*Order, error)
	GetOrdersByUser(ctx context.Context, userID string, limit, offset int) ([]*Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status OrderStatus) error
	CancelOrder(ctx context.Context, id string) error
	ListOrders(ctx context.Context, limit, offset int) ([]*Order, error)
}
