package service

import (
	"context"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
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

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if err := utils.ValidateStruct(product); err != nil {
		return nil, err
	}

	newProduct, err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	// Publish event
	return newProduct, s.nats.PublishProductCreated(product)
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

func (s *ProductServiceImpl) UpdateProduct(ctx context.Context, id string, product *domain.Product) (*domain.Product, error) {
	if err := utils.ValidateStruct(product); err != nil {
		return nil, err
	}

	updatedProduct, err := s.repo.Update(ctx, id, product)
	if err != nil {
		return nil, err
	}

	return updatedProduct, s.nats.PublishProductUpdated(product)
}

func (s *ProductServiceImpl) DeleteProduct(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return nil
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
		if errors.Is(err, repository.ErrProductNotFound) {
			return false, -1, ErrProductNotFound
		}
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
		return ErrInsufficientStock
	}

	if err := s.repo.UpdateStock(ctx, id, -quantity); err != nil {
		return err
	}

	return s.nats.PublishStockUpdated(id, n, quantity)
}

func (s *ProductServiceImpl) RestoreStock(ctx context.Context, id string, quantity int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, quantity); err != nil {
		return err
	}

	return s.nats.PublishStockUpdated(id, product.Stock, quantity)
}
