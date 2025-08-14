package service

import (
	"context"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

var (
	ErrOrderNotFound                = errors.New("Order not found")
	ErrInvalidOrderStatusTransition = errors.New("Invalid Order status transition")
	ErrOrderStateChanged            = errors.New("Order status has changed, please try again")
	ErrOrderCannotBeCancelled       = errors.New("Order cannot be cancled, it is too late...")
)

type OrderServiceImpl struct {
	repo domain.OrderRepository
	nats *messaging.OrderEventPublisher
}

func NewOrderService(repo domain.OrderRepository, nats *messaging.OrderEventPublisher) domain.OrderService {
	service := &OrderServiceImpl{
		repo: repo,
		nats: nats,
	}

	return service
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	if err := utils.ValidateStruct(order); err != nil {
		return nil, err
	}

	// Calculate total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	// Check stock availability for all items
	if err := s.checkStockAvailability(order.Items); err != nil {
		return nil, err
	}

	// Create order
	newOrder, err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Reserve stock for all items
	s.reserveStock(order.Items)

	if err := s.nats.PublishOrderCreated(order); err != nil {
		return nil, err
	}

	return newOrder, err
}

func (s *OrderServiceImpl) checkStockAvailability(items []domain.OrderItem) error {
	for _, item := range items {
		// Publish stock check request
		s.nats.PublishStockCheck(&item)
	}
	// TODO:wait for response
	return nil
}

func (s *OrderServiceImpl) reserveStock(items []domain.OrderItem) {
	for _, item := range items {
		s.nats.PublishReserveStock(&item)
	}
}

func (s *OrderServiceImpl) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

func (s *OrderServiceImpl) GetOrdersByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

func (s *OrderServiceImpl) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		return nil, ErrInvalidOrderStatusTransition
	}

	updatedOrder, err := s.repo.UpdateStatus(ctx, id, order.Status, status)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderStateChanged
		}
		return nil, err
	}

	// Publish order updated event
	if err := s.nats.PublishOrderUpdated(updatedOrder, status); err != nil {
		return nil, err
	}

	return updatedOrder, nil
}

func (s *OrderServiceImpl) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.UpdateOrderStatus(ctx, id, domain.OrderStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrInvalidOrderStatusTransition) {
			return nil, ErrOrderCannotBeCancelled
		}
		return nil, err
	}
	return order, nil
}

func (s *OrderServiceImpl) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *OrderServiceImpl) isValidStatusTransition(current, new domain.OrderStatus) bool {
	validTransitions := map[domain.OrderStatus][]domain.OrderStatus{
		domain.OrderStatusPending:   {domain.OrderStatusConfirmed, domain.OrderStatusCancelled},
		domain.OrderStatusConfirmed: {domain.OrderStatusShipped, domain.OrderStatusCancelled},
		domain.OrderStatusShipped:   {domain.OrderStatusDelivered},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}
