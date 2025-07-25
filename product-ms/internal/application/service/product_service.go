package service

import (
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messageing"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type ProductServiceImpl struct {
	repo domain.ProductRepository
	nats *messaging.ProductEventPublisher
}

func NewProductService(repo domain.ProductRepository, nats *messaging.ProductEventPublisher) domain.ProductService {
	return &ProductServiceImpl{
		repo: repo,
		nats: nats,
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
	return s.nats.PublishProductCreated(product)
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

	return s.nats.PublishProductUpdated(product)
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
	hasStock, n, err := s.CheckStock(id, quantity)
	if err != nil {
		return err
	}

	if !hasStock {
		return errors.New("insufficient stock")
	}

	s.repo.UpdateStock(id, quantity)

	return s.nats.PublishStockUpdated(id, n, quantity)
}
