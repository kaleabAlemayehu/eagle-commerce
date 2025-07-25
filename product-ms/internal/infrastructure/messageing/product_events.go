// services/product-ms/internal/infrastructure/messaging/product_events.go
package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
)

type ProductEventPublisher struct {
	natsClient *messaging.NATSClient
}

func NewProductEventPublisher(natsClient *messaging.NATSClient) *ProductEventPublisher {
	return &ProductEventPublisher{
		natsClient: natsClient,
	}
}

func (p *ProductEventPublisher) PublishProductCreated(product *domain.Product) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   models.ProductCreatedEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": product.ID.Hex(),
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
			"category":   product.Category,
			"active":     product.Active,
			"created_at": product.CreatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("product.created", event)
}

func (p *ProductEventPublisher) PublishProductUpdated(product *domain.Product) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   models.ProductUpdatedEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": product.ID.Hex(),
			"name":       product.Name,
			"price":      product.Price,
			"stock":      product.Stock,
			"category":   product.Category,
			"active":     product.Active,
			"updated_at": product.UpdatedAt,
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("product.updated", event)
}

func (p *ProductEventPublisher) PublishStockUpdated(productID string, oldStock, newStock int) error {
	event := models.Event{
		ID:     generateEventID(),
		Type:   "product.stock.updated",
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": productID,
			"old_stock":  oldStock,
			"new_stock":  newStock,
			"updated_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish("product.stock.updated", event)
}

type ProductEventHandler struct {
	productService domain.ProductService
	natsClient     *messaging.NATSClient
}

func NewProductEventHandler(productService domain.ProductService, natsClient *messaging.NATSClient) *ProductEventHandler {
	return &ProductEventHandler{
		productService: productService,
		natsClient:     natsClient,
	}
}

func (h *ProductEventHandler) StartListening() error {
	// Subscribe to stock-related events
	_, err := h.natsClient.Subscribe("stock.check", h.handleStockCheck)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe("stock.reserve", h.handleStockReserve)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe("order.cancelled", h.handleOrderCancelled)
	return err
}

func (h *ProductEventHandler) handleStockCheck(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling stock.check event: %v", err)
		return
	}

	productID, ok := event.Data["product_id"].(string)
	if !ok {
		log.Printf("Invalid product_id in stock.check event")
		return
	}

	quantity, ok := event.Data["quantity"].(float64)
	if !ok {
		log.Printf("Invalid quantity in stock.check event")
		return
	}

	available, _, err := h.productService.CheckStock(productID, int(quantity))
	if err != nil {
		log.Printf("Error checking stock: %v", err)
		return
	}

	// Publish response
	responseEvent := models.Event{
		ID:     generateEventID(),
		Type:   "stock.check.response",
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": productID,
			"quantity":   int(quantity),
			"available":  available,
			"request_id": event.ID,
		},
		Timestamp: time.Now(),
	}

	h.natsClient.Publish("stock.check.response", responseEvent)
}

func (h *ProductEventHandler) handleStockReserve(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling stock.reserve event: %v", err)
		return
	}

	productID, ok := event.Data["product_id"].(string)
	if !ok {
		log.Printf("Invalid product_id in stock.reserve event")
		return
	}

	quantity, ok := event.Data["quantity"].(float64)
	if !ok {
		log.Printf("Invalid quantity in stock.reserve event")
		return
	}

	err := h.productService.ReserveStock(productID, int(quantity))
	if err != nil {
		log.Printf("Error reserving stock: %v", err)
		return
	}

	log.Printf("Stock reserved for product %s: %d units", productID, int(quantity))
}

func (h *ProductEventHandler) handleOrderCancelled(data []byte) {
	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Error unmarshaling order.cancelled event: %v", err)
		return
	}

	items, ok := event.Data["items"].([]interface{})
	if !ok {
		log.Printf("Invalid items in order.cancelled event")
		return
	}

	// Restore stock for cancelled order items
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		productID, ok := itemMap["product_id"].(string)
		if !ok {
			continue
		}

		quantity, ok := itemMap["quantity"].(float64)
		if !ok {
			continue
		}

		// TODO:
		// Restore stock by adding back the quantity
		// This would need a RestoreStock method in the service
		log.Printf("Restoring stock for product %s: %d units", productID, int(quantity))
	}
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + "product"
}
