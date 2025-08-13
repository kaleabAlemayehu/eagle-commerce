package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/application/service"
	messaging "github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedLogger "github.com/kaleabAlemayehu/eagle-commerce/shared/logger"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
)

// @title Product Service API
// @version 1.0
// @description This is a product service API for eCommerce application
// @host localhost:8082
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	// Create logger
	logger := sharedLogger.NewLogger().With("Service", "Product")

	// Connect to MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		logger.Error("Failed to connect to MongoDB:", "error", err)
		return
	}

	// Connect to NATS
	natsClient, err := sharedMessaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		logger.Error("Failed to connect to NATS:", "error", err)
		return
	}
	nats := messaging.NewProductEventPublisher(natsClient)

	// Initialize dependencies
	productRepo := repository.NewMongoProductRepository(db.Database)
	productService := service.NewProductService(productRepo, nats)
	productHandler := handler.NewProductHandler(productService)
	if err = messaging.NewProductEventHandler(productService, natsClient).StartListening(); err != nil {
		logger.Error("Failed to listen events: ", "error", err)
		return
	}

	// Setup router
	r := router.NewRouter(productHandler, logger)

	port := "8082"
	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		logger.Info("Product service starting on port: " + port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("ListenAndServer error ", "error", err)
		}

	}()
	<-stop

	logger.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server Shutdown error", "error", err)
	}
	natsClient.Close()

	if err := db.Close(); err != nil {
		logger.Error("Error closing MongoDB connection", "error", err)
	}

	logger.Info("Server stopped")

}
