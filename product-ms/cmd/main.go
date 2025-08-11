package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/application/service"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messageing"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
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

	// Connect to NATS
	natsClient, err := sharedMessaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
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

	port := "8082"
	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		fmt.Printf("Product service starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServer error err: %v", err)
		}

	}()
	<-stop

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
	}
	natsClient.Close()

	if err := db.Close(); err != nil {
		log.Printf("Error closing MongoDB connection: %v", err)
	}

	log.Println("Server stopped")

}
