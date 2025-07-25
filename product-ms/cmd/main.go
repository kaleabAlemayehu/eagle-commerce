package main

import (
	"fmt"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/application/service"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messageing"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"log"
	"net/http"
)

// @title Product Service API
// @version 1.0
// @description This is a product service API for eCommerce application
// @host localhost:8082
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
	nats := messaging.NewProductEventPublisher(natsClient)

	// Initialize dependencies
	productRepo := repository.NewMongoProductRepository(db.Database)
	productService := service.NewProductService(productRepo, nats)
	productHandler := handler.NewProductHandler(productService)
	if err = messaging.NewProductEventHandler(productService, natsClient).StartListening(); err != nil {
		log.Fatal("Failed to listen events: ", err)
	}

	// Setup router
	r := router.NewRouter(productHandler)

	// Start server
	port := "8082"
	fmt.Printf("Product service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
