package service

import (
	"encoding/json"
	"errors"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type OrderServiceImpl struct {
	repo       domain.OrderRepository
	natsClient *messaging.NATSClient
}

func NewOrderService(repo domain.OrderRepository, natsClient *messaging.NATSClient) domain.OrderService {
	service := &OrderServiceImpl{
		repo:       repo,
		natsClient: natsClient,
	}

	// Subscribe to events
	service.subscribeToEvents()

	return service
}

func (s *OrderServiceImpl) subscribeToEvents() {
	// Subscribe to payment events
	s.natsClient.Subscribe("payment.processed", s.handlePaymentProcessed)
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

	// Publish order created event
	event := models.Event{
		Type:   models.OrderCreatedEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id": order.ID.Hex(),
			"user_id":  order.UserID,
			"total":    order.Total,
			"items":    order.Items,
		},
	}
	s.natsClient.Publish("order.created", event)

	return nil
}

func (s *OrderServiceImpl) checkStockAvailability(items []domain.OrderItem) error {
	for _, item := range items {
		// Publish stock check request
		event := models.Event{
			Type:   "stock.check",
			Source: "order-service",
			Data: map[string]interface{}{
				"product_id": item.ProductID,
				"quantity":   item.Quantity,
			},
		}
		s.natsClient.Publish("stock.check", event)
	}
	return nil // In real implementation, wait for response
}

func (s *OrderServiceImpl) reserveStock(items []domain.OrderItem) {
	for _, item := range items {
		event := models.Event{
			Type:   "stock.reserve",
			Source: "order-service",
			Data: map[string]interface{}{
				"product_id": item.ProductID,
				"quantity":   item.Quantity,
			},
		}
		s.natsClient.Publish("stock.reserve", event)
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
	event := models.Event{
		Type:   models.OrderUpdatedEvent,
		Source: "order-service",
		Data: map[string]interface{}{
			"order_id": id,
			"status":   status,
			"user_id":  order.UserID,
		},
	}
	s.natsClient.Publish("order.updated", event)

	return nil
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
