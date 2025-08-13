package service

import (
	"context"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/logger"
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

	logger := logger.FromContext(ctx).With("Layer", "service", "method", "CreateProduct")
	if err != nil {
		logger.Error("Repository Failed to Create Product", "error", err)
		return nil, err
	}

	// Publish event
	if err := s.nats.PublishProductCreated(product); err != nil {
		logger.Error("NATS Failed to Publish ProductCreated", "error", err)
		return nil, err
	}
	logger.Info("Product created successfully")
	return newProduct, nil
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "GetProduct", "product_id", id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			logger.Warn("Product not Found on repository")
			return nil, ErrProductNotFound
		}
		logger.Error("Failed to retrieve product from the repository", "error", err)
		return nil, err
	}
	logger.Info("Product retrived successfully")
	return product, nil
}

func (s *ProductServiceImpl) UpdateProduct(ctx context.Context, id string, product *domain.Product) (*domain.Product, error) {
	if err := utils.ValidateStruct(product); err != nil {
		return nil, err
	}
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "UpdateProduct", "product_id", id)
	updatedProduct, err := s.repo.Update(ctx, id, product)
	if err != nil {
		logger.Error("Repository Failed to Update Product", "error", err)
		return nil, err
	}

	if err := s.nats.PublishProductUpdated(product); err != nil {

		logger.Error("NATS Failed to Publish ProductUpdated", "error", err)
		return nil, err
	}

	logger.Info("Product updated succesfully")
	return updatedProduct, nil
}

func (s *ProductServiceImpl) DeleteProduct(ctx context.Context, id string) error {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "Product", "product_id", id)
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			logger.Warn("Product not Found on repository")
			return ErrProductNotFound
		}
		logger.Error("Repository Failed to Delete Product", "error", err)
		return err
	}
	logger.Info("Product got deleted succefully")
	return nil
}

func (s *ProductServiceImpl) ListProducts(ctx context.Context, limit, offset int, category string) ([]*domain.Product, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "ListProducts", "limit", limit, "offset", offset, "category", category)
	products, err := s.repo.List(ctx, limit, offset, category)
	if err != nil {
		logger.Error("Failed to list products from repository", "error", err)
		return nil, err
	}
	logger.Info("Products listed successfully")
	return products, nil
}

func (s *ProductServiceImpl) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "SearchProducts", "query", query, "limit", limit, "offset", offset)
	products, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		logger.Error("Failed to search products in repository", "error", err)
		return nil, err
	}
	logger.Info("Products searched successfully")
	return products, nil
}

func (s *ProductServiceImpl) CheckStock(ctx context.Context, id string, quantity int) (bool, int, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "CheckStock", "product_id", id, "quantity", quantity)
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			logger.Warn("Product not found for stock check")
			return false, -1, ErrProductNotFound
		}
		logger.Error("Failed to get product for stock check", "error", err)
		return false, -1, err
	}

	hasStock := product.Stock >= quantity
	logger.Info("Stock checked successfully", "has_stock", hasStock, "current_stock", product.Stock)
	return hasStock, product.Stock, nil
}

func (s *ProductServiceImpl) ReserveStock(ctx context.Context, id string, quantity int) error {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "ReserveStock", "product_id", id, "quantity", quantity)
	hasStock, n, err := s.CheckStock(ctx, id, quantity)
	if err != nil {
		logger.Error("Failed to check stock during reservation", "error", err)
		return err
	}

	if !hasStock {
		logger.Warn("Insufficient stock for reservation")
		return ErrInsufficientStock
	}

	if err := s.repo.UpdateStock(ctx, id, -quantity); err != nil {
		logger.Error("Failed to update stock in repository", "error", err)
		return err
	}

	if err := s.nats.PublishStockUpdated(id, n, quantity); err != nil {
		logger.Error("NATS Failed to Publish StockUpdated", "error", err)
		return err
	}

	logger.Info("Stock reserved successfully")
	return nil
}

func (s *ProductServiceImpl) RestoreStock(ctx context.Context, id string, quantity int) error {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "RestoreStock", "product_id", id, "quantity", quantity)
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			logger.Warn("Product not found for stock restore")
			return ErrProductNotFound
		}
		logger.Error("Failed to get product for stock restore", "error", err)
		return err
	}

	if err := s.repo.UpdateStock(ctx, id, quantity); err != nil {
		logger.Error("Failed to update stock in repository", "error", err)
		return err
	}

	if err := s.nats.PublishStockUpdated(id, product.Stock, quantity); err != nil {
		logger.Error("NATS Failed to Publish StockUpdated", "error", err)
		return err
	}

	logger.Info("Stock restored successfully")
	return nil
}
