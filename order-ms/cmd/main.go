package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
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
	natsClient, err := messaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsClient.Close()

	// Initialize dependencies
	orderRepo := repository.NewMongoOrderRepository(db.Database)
	orderService := service.NewOrderService(orderRepo, natsClient)
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup router
	r := router.NewRouter(orderHandler)

	// Start server
	port := "8083"
	fmt.Printf("Order service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
