package messaging

import (
	"context"
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
		ID:     messaging.GenerateEventID(),
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

	return p.natsClient.Publish(models.ProductCreatedEvent, event)
}

func (p *ProductEventPublisher) PublishProductUpdated(product *domain.Product) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
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

	return p.natsClient.Publish(models.ProductUpdatedEvent, event)
}

func (p *ProductEventPublisher) PublishStockUpdated(productID string, oldStock, newStock int) error {
	event := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.ProductStockUpdatedEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": productID,
			"old_stock":  oldStock,
			"new_stock":  newStock,
			"updated_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return p.natsClient.Publish(models.ProductStockUpdatedEvent, event)
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
	_, err := h.natsClient.Subscribe(models.StockCheckEvent, h.handleStockCheck)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe(models.StockReserveEvent, h.handleStockReserve)
	if err != nil {
		return err
	}

	_, err = h.natsClient.Subscribe(models.OrderCancelledEvent, h.handleOrderCancelled)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	available, _, err := h.productService.CheckStock(ctx, productID, int(quantity))
	if err != nil {
		log.Printf("Error checking stock: %v", err)
		return
	}

	// Publish response
	responseEvent := models.Event{
		ID:     messaging.GenerateEventID(),
		Type:   models.StockCheckResponseEvent,
		Source: "product-service",
		Data: map[string]interface{}{
			"product_id": productID,
			"quantity":   int(quantity),
			"available":  available,
			"request_id": event.ID,
		},
		Timestamp: time.Now(),
	}

	h.natsClient.Publish(models.StockCheckResponseEvent, responseEvent)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := h.productService.ReserveStock(ctx, productID, int(quantity))
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.productService.RestoreStock(ctx, productID, int(quantity)); err != nil {
			log.Printf("Error restoring stock: %v", err)
			return
		}

		log.Printf("Restoring stock for product %s: %d units", productID, int(quantity))
	}
}
