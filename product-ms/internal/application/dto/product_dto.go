package dto

import "time"

type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Stock       int      `json:"stock" validate:"gte=0"`
	Category    string   `json:"category" validate:"required"`
	Images      []string `json:"images"`
}

type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock,omitempty" validate:"omitempty,gte=0"`
	Category    string   `json:"category,omitempty"`
	Images      []string `json:"images,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	Images      []string  `json:"images"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
}

type ProductSearchResponse struct {
	Products []ProductResponse `json:"products"`
	Query    string            `json:"query"`
	Total    int               `json:"total"`
}

type StockUpdateRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=add subtract set"`
}

type StockCheckRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"gt=0"`
}

type StockCheckResponse struct {
	ProductID string `json:"product_id"`
	Available bool   `json:"available"`
	Stock     int    `json:"current_stock"`
	Requested int    `json:"requested_quantity"`
}

type CategoryResponse struct {
	Categories []string `json:"categories"`
	Total      int      `json:"total"`
}
