package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name" validate:"required"`
	Description string             `json:"description" bson:"description"`
	Price       float64            `json:"price" bson:"price" validate:"required,gt=0"`
	Stock       int                `json:"stock" bson:"stock" validate:"gte=0"`
	Category    string             `json:"category" bson:"category" validate:"required"`
	Images      []string           `json:"images" bson:"images"`
	Active      bool               `json:"active" bson:"active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, id string, product *Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, category string) ([]*Product, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Product, error)
	UpdateStock(ctx context.Context, id string, quantity int) error
}

type ProductService interface {
	CreateProduct(ctx context.Context, product *Product) error
	GetProduct(ctx context.Context, id string) (*Product, error)
	UpdateProduct(ctx context.Context, id string, product *Product) error
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, limit, offset int, category string) ([]*Product, error)
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]*Product, error)
	CheckStock(ctx context.Context, id string, quantity int) (bool, int, error)
	ReserveStock(ctx context.Context, id string, quantity int) error
	RestoreStock(ctx context.Context, id string, quantity int) error
}
