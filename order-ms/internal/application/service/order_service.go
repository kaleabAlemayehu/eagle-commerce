package service

import (
	"encoding/json"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
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

func (s *OrderServiceImpl) handlePaymentProcessed(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return
	}

	status, ok := event.Data["status"].(string)
	if !ok {
		return
	}

	if status == "completed" {
		s.UpdateOrderStatus(orderID, domain.OrderStatusConfirmed)
	} else if status == "failed" {
		s.UpdateOrderStatus(orderID, domain.OrderStatusCancelled)
	}
}

func (s *OrderServiceImpl) CreateOrder(order *domain.Order) error {
	if err := utils.ValidateStruct(order); err != nil {
		return err
	}

	// Calculate total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	// Check stock availability for all items
	if err := s.checkStockAvailability(order.Items); err != nil {
		return err
	}

	// Create order
	if err := s.repo.Create(order); err != nil {
		return err
	}

	// Reserve stock for all items
	s.reserveStock(order.Items)

	return s.nats.PublishOrderCreated(order)
}

func (s *OrderServiceImpl) checkStockAvailability(items []domain.OrderItem) error {
	for _, item := range items {
		// Publish stock check request
		s.nats.PublishStockCheck(&item)
	}
	return nil // In real implementation, wait for response
}

func (s *OrderServiceImpl) reserveStock(items []domain.OrderItem) {
	for _, item := range items {
		s.nats.PublishReserveStock(&item)
	}
}

func (s *OrderServiceImpl) GetOrder(id string) (*domain.Order, error) {
	return s.repo.GetByID(id)
}

func (s *OrderServiceImpl) GetOrdersByUser(userID string, limit, offset int) ([]*domain.Order, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *OrderServiceImpl) UpdateOrderStatus(id string, status domain.OrderStatus) error {
	order, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		return errors.New("invalid status transition")
	}

	if err := s.repo.UpdateStatus(id, status); err != nil {
		return err
	}

	// Publish order updated event

	return s.nats.PublishOrderUpdated(order, status)
}

func (s *OrderServiceImpl) CancelOrder(id string) error {
	return s.UpdateOrderStatus(id, domain.OrderStatusCancelled)
}

func (s *OrderServiceImpl) ListOrders(limit, offset int) ([]*domain.Order, error) {
	return s.repo.List(limit, offset)
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
