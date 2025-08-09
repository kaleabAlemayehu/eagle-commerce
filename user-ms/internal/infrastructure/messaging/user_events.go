package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type UserEventPublisher struct {
	natsClient *messaging.NATSClient
}

func NewUserEventPublisher(natsClient *messaging.NATSClient) *UserEventPublisher {
	return &UserEventPublisher{
		natsClient: natsClient,
	}
}

func (p *UserEventPublisher) PublishUserCreated(user *domain.User) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   models.UserCreatedEvent,
		Source: "user-service",
		Data: map[string]interface{}{
			"user_id":    user.ID.Hex(),
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.UserCreatedEvent, event)
}

func (p *UserEventPublisher) PublishUserUpdated(user *domain.User) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   models.UserUpdatedEvent,
		Source: "user-service",
		Data: map[string]interface{}{
			"user_id":    user.ID.Hex(),
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"updated_at": user.UpdatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.UserUpdatedEvent, event)
}

func (p *UserEventPublisher) PublishUserDeleted(userID string) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   models.UserDeletedEvent,
		Source: "user-service",
		Data: map[string]interface{}{
			"user_id":    userID,
			"deleted_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.UserDeletedEvent, event)
}

type UserEventHandler struct {
	natsClient *messaging.NATSClient
}

func NewUserEventHandler(natsClient *messaging.NATSClient) *UserEventHandler {
	return &UserEventHandler{
		natsClient: natsClient,
	}
}

func (h *UserEventHandler) StartListening() error {
	// Subscribe to user-related events from other services
	_, err := h.natsClient.Subscribe(models.OrderCreatedEvent, h.handleOrderCreated)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe(models.PaymentProcessedEvent, h.handlePaymentProcessed)
	return err
}

func (h *UserEventHandler) handleOrderCreated(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling order.created event: %v", err)
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Invalid user_id in order.created event")
		return
	}

	log.Printf("User %s created a new order: %s", userID, event.Data["order_id"])

	// INFO:
	// Here I could update user statistics, send notifications, etc.
}

func (h *UserEventHandler) handlePaymentProcessed(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling payment.processed event: %v", err)
		return
	}

	// INFO:
	// Handle payment processing for user notifications, etc.
	log.Printf("Payment processed: %+v", event.Data)
}

func generateEventID() string {
	return uuid.NewString()
}
