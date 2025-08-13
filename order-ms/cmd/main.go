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

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedLogger "github.com/kaleabAlemayehu/eagle-commerce/shared/logger"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
)

// @title Order Service API
// @version 1.0
// @description This is an order service API for eCommerce application
// @host localhost:8083
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	// create logerlogger
	logger := sharedLogger.NewLogger()

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// Start server
	port := "8083"
	fmt.Printf("Order service starting on port %s\n", port)
	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
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
