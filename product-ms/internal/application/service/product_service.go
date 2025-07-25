package service

import (
	"errors"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type ProductServiceImpl struct {
	repo       domain.ProductRepository
	natsClient *messaging.NATSClient
}

func NewProductService(repo domain.ProductRepository, natsClient *messaging.NATSClient) domain.ProductService {
	return &ProductServiceImpl{
		repo:       repo,
		natsClient: natsClient,
	}
}

func (s *ProductServiceImpl) CreateProduct(product *domain.Product) error {
	if err := utils.ValidateStruct(product); err != nil {
		return err
	}

	if err := s.repo.Create(product); err != nil {
		return err
	}

	// Publish event
	event := models.Event{
		Type:   models.ProductCreatedEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": product.ID.Hex(),
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
		},
	}
	s.natsClient.Publish("product.created", event)

	return nil
}

func (s *ProductServiceImpl) GetProduct(id string) (*domain.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductServiceImpl) UpdateProduct(id string, product *domain.Product) error {
	if err := utils.ValidateStruct(product); err != nil {
		return err
	}

	if err := s.repo.Update(id, product); err != nil {
		return err
	}

	// Publish event
	event := models.Event{
		Type:   models.ProductUpdatedEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": id,
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
		},
	}
	s.natsClient.Publish("product.updated", event)

	return nil
}

func (s *ProductServiceImpl) DeleteProduct(id string) error {
	return s.repo.Delete(id)
}

func (s *ProductServiceImpl) ListProducts(limit, offset int, category string) ([]*domain.Product, error) {
	return s.repo.List(limit, offset, category)
}

func (s *ProductServiceImpl) SearchProducts(query string, limit, offset int) ([]*domain.Product, error) {
	return s.repo.Search(query, limit, offset)
}

func (s *ProductServiceImpl) CheckStock(id string, quantity int) (bool, int, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return false, -1, err
	}

	return product.Stock >= quantity, product.Stock, nil
}

func (s *ProductServiceImpl) ReserveStock(id string, quantity int) error {
	hasStock, _, err := s.CheckStock(id, quantity)
	if err != nil {
		return err
	}

	if !hasStock {
		return errors.New("insufficient stock")
	}

	return s.repo.UpdateStock(id, quantity)
}
