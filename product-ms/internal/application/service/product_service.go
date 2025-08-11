package service

import (
	"context"
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

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, product *domain.Product) error {
	if err := utils.ValidateStruct(product); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return err
	}

	// Publish event
	return s.nats.PublishProductCreated(product)
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductServiceImpl) UpdateProduct(ctx context.Context, id string, product *domain.Product) error {
	if err := utils.ValidateStruct(product); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, id, product); err != nil {
		return err
	}

	return s.nats.PublishProductUpdated(product)
}

func (s *ProductServiceImpl) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductServiceImpl) ListProducts(ctx context.Context, limit, offset int, category string) ([]*domain.Product, error) {
	return s.repo.List(ctx, limit, offset, category)
}

func (s *ProductServiceImpl) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
	return s.repo.Search(ctx, query, limit, offset)
}

func (s *ProductServiceImpl) CheckStock(ctx context.Context, id string, quantity int) (bool, int, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, -1, err
	}

	return product.Stock >= quantity, product.Stock, nil
}

func (s *ProductServiceImpl) ReserveStock(ctx context.Context, id string, quantity int) error {
	hasStock, n, err := s.CheckStock(ctx, id, quantity)
	if err != nil {
		return err
	}

	if !hasStock {
		return errors.New("insufficient stock")
	}

	if err := s.repo.UpdateStock(ctx, id, -quantity); err != nil {
		return err
	}

	return s.nats.PublishStockUpdated(id, n, quantity)
}

func (s *ProductServiceImpl) RestoreStock(ctx context.Context, id string, quantity int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, quantity); err != nil {
		return err
	}

	return s.nats.PublishStockUpdated(id, product.Stock, quantity)
}
