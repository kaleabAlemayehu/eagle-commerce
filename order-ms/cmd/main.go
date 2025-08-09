package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
)

// @title Order Service API
// @version 1.0
// @description This is an order service API for eCommerce application
// @host localhost:8083
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	// Connect to MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	// Connect to NATS
	natsClient, err := sharedMessaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsClient.Close()
	nats := messaging.NewOrderEventPublisher(natsClient)

	// Initialize dependencies
	orderRepo := repository.NewMongoOrderRepository(db.Database)
	orderService := service.NewOrderService(orderRepo, nats)
	orderHandler := handler.NewOrderHandler(orderService)
	if err := messaging.NewOrderEventHandler(orderService, natsClient).StartListening(); err != nil {
		log.Fatal("Failed to listen NATS events:", err)
	}
	mode := cfg.Environment

	// Setup router
	r := router.NewRouter(orderHandler, mode)

	// Start server
	port := "8083"
	fmt.Printf("Order service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
