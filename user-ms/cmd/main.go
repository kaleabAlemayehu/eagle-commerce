package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

// @title User Service API
// @version 1.0
// @description This is a user service API for eCommerce application
// @host localhost:8081
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
	nats := messaging.NewUserEventPublisher(natsClient)

	// listening for the incoming events from other services
	messaging.NewUserEventHandler(natsClient).StartListening()

	// Initialize dependencies
	userRepo := repository.NewMongoUserRepository(db.Database)
	auth := sharedMiddleware.NewAuth(cfg.JWTSecret)
	userService := service.NewUserService(userRepo, nats, auth)
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	r := router.NewRouter(userHandler, auth)

	// Start server
	port := "8081"
	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("User service starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe Error: %v ", err)
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
