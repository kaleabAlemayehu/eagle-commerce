package service

import (
	"context"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/logger"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

var (
	ErrOrderNotFound                = errors.New("Order not found")
	ErrInvalidOrderStatusTransition = errors.New("Invalid Order status transition")
	ErrOrderStateChanged            = errors.New("Order status has changed, please try again")
	ErrOrderCannotBeCancelled       = errors.New("Order cannot be cancled, it is too late...")
	ErrOrderOutOfStock              = errors.New("Order out of stock.")
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
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "CreateOrder", "user_id", order.UserID)
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
	if isAvailable := s.checkStockAvailability(order.Items); !isAvailable {
		logger.Warn("order out of stock", "order_items", order.Items)
		return nil, ErrOrderOutOfStock
	}

	// Create order
	newOrder, err := s.repo.Create(ctx, order)
	if err != nil {
		logger.Error("failed to create order in repository", "error", err)
		return nil, err
	}

	// Reserve stock for all items
	s.reserveStock(order.Items)

	if err := s.nats.PublishOrderCreated(order); err != nil {
		logger.Error("failed to publish order created event", "error", err)
		return nil, err
	}

	logger.Info("order created successfully", "order_id", newOrder.ID)
	return newOrder, err
}

func (s *OrderServiceImpl) checkStockAvailability(items []domain.OrderItem) bool {
	isAllAvailable := true
	for _, item := range items {
		// using request instead of publish
		isAvaliable, err := s.nats.RequestStockCheck(&item)
		if err != nil || !isAvaliable {
			isAllAvailable = false
			break
		}
	}
	return isAllAvailable
}

func (s *OrderServiceImpl) reserveStock(items []domain.OrderItem) {
	for _, item := range items {
		s.nats.PublishReserveStock(&item)
	}
}

func (s *OrderServiceImpl) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "GetOrder", "order_id", id)
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			logger.Warn("order not found in repository")
			return nil, ErrOrderNotFound
		}
		logger.Error("failed to get order from repository", "error", err)
		return nil, err
	}
	logger.Info("order retrieved successfully")
	return order, nil
}

func (s *OrderServiceImpl) GetOrdersByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "GetOrdersByUser", "user_id", userID, "limit", limit, "offset", offset)
	orders, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("failed to get orders by user from repository", "error", err)
		return nil, err
	}
	logger.Info("orders retrieved successfully for user")
	return orders, nil
}

func (s *OrderServiceImpl) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "UpdateOrderStatus", "order_id", id, "new_status", status)
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			logger.Warn("order not found in repository")
			return nil, ErrOrderNotFound
		}
		logger.Error("failed to get order for status update from repository", "error", err)
		return nil, err
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		logger.Warn("invalid order status transition", "current_status", order.Status)
		return nil, ErrInvalidOrderStatusTransition
	}

	updatedOrder, err := s.repo.UpdateStatus(ctx, id, order.Status, status)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			logger.Warn("order status changed during update")
			return nil, ErrOrderStateChanged
		}
		logger.Error("failed to update order status in repository", "error", err)
		return nil, err
	}

	// Publish order updated event
	if err := s.nats.PublishOrderUpdated(updatedOrder, status); err != nil {
		logger.Error("failed to publish order updated event", "error", err)
		return nil, err
	}

	logger.Info("order status updated successfully")
	return updatedOrder, nil
}

func (s *OrderServiceImpl) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "CancelOrder", "order_id", id)
	logger.Info("attempting to cancel order")
	order, err := s.UpdateOrderStatus(ctx, id, domain.OrderStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrInvalidOrderStatusTransition) {
			logger.Warn("order cannot be cancelled in its current state")
			return nil, ErrOrderCannotBeCancelled
		}
		return nil, err
	}
	logger.Info("order cancelled successfully")
	return order, nil
}

func (s *OrderServiceImpl) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	logger := logger.FromContext(ctx).With("Layer", "service", "method", "ListOrders", "limit", limit, "offset", offset)
	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		logger.Error("failed to list orders from repository", "error", err)
		return nil, err
	}
	logger.Info("orders listed successfully")
	return orders, nil
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
