package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
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
	Create(product *Product) error
	GetByID(id string) (*Product, error)
	Update(id string, product *Product) error
	Delete(id string) error
	List(limit, offset int, category string) ([]*Product, error)
	Search(query string, limit, offset int) ([]*Product, error)
	UpdateStock(id string, quantity int) error
}

type ProductService interface {
	CreateProduct(product *Product) error
	GetProduct(id string) (*Product, error)
	UpdateProduct(id string, product *Product) error
	DeleteProduct(id string) error
	ListProducts(limit, offset int, category string) ([]*Product, error)
	SearchProducts(query string, limit, offset int) ([]*Product, error)
	CheckStock(id string, quantity int) (bool, error)
	ReserveStock(id string, quantity int) error
}
